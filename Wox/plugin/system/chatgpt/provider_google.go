package chatgpt

import (
	"context"
	"errors"
	"github.com/google/generative-ai-go/genai"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"wox/plugin"
)

type GoogleProvider struct {
	connectContext chatgptProviderConnectContext
	client         *genai.Client
	api            plugin.API
}

type GoogleProviderStream struct {
	stream        *genai.GenerateContentResponseIterator
	conversations []Conversation
}

func NewGoogleProvider(ctx context.Context, connectContext chatgptProviderConnectContext, api plugin.API) Provider {
	return &GoogleProvider{connectContext: connectContext, api: api}
}

func (g *GoogleProvider) Connect(ctx context.Context) error {
	client, newClientErr := genai.NewClient(ctx, option.WithAPIKey(g.connectContext.ApiKey))
	if newClientErr != nil {
		return newClientErr
	}

	g.client = client
	return nil
}

func (g *GoogleProvider) Close(ctx context.Context) error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func (g *GoogleProvider) ChatStream(ctx context.Context, model chatgptModel, conversations []Conversation) (ProviderChatStream, error) {
	chatMessages, lastConversation := g.convertConversations(conversations)
	aiModel := g.client.GenerativeModel(model.Name)
	session := aiModel.StartChat()
	session.History = chatMessages
	stream := session.SendMessageStream(ctx, lastConversation.Parts...)
	return &GoogleProviderStream{conversations: conversations, stream: stream}, nil
}

func (g *GoogleProvider) Chat(ctx context.Context, model chatgptModel, conversations []Conversation) (string, error) {
	chatMessages, lastConversation := g.convertConversations(conversations)
	aiModel := g.client.GenerativeModel(model.Name)
	session := aiModel.StartChat()
	session.History = chatMessages
	response, sendErr := session.SendMessage(ctx, lastConversation.Parts...)
	if sendErr != nil {
		return "", sendErr
	}

	for _, part := range response.Candidates[0].Content.Parts {
		if v, ok := part.(genai.Text); ok {
			return string(v), nil
		}
	}

	return "", errors.New("no text in response")
}

func (g *GoogleProvider) Models(ctx context.Context) ([]chatgptModel, error) {
	return []chatgptModel{
		{
			DisplayName: "google-gemini-1.0-pro",
			Name:        "gemini-1.0-pro",
			Provider:    chatgptModelProviderNameGoogle,
		},
	}, nil
}

func (g *GoogleProviderStream) Receive(ctx context.Context) (string, error) {
	response, err := g.stream.Next()
	if err != nil {
		// no more messages
		if errors.Is(err, iterator.Done) {
			return "", io.EOF
		}

		var v *apierror.APIError
		if errors.As(err, &v) {
			return "", v.Unwrap()
		}

		return "", err
	}
	if len(response.Candidates) == 0 {
		return "", io.EOF
	}

	for _, part := range response.Candidates[0].Content.Parts {
		if v, ok := part.(genai.Text); ok {
			return string(v), nil
		}
	}

	return "", errors.New("no text in response")
}

func (g *GoogleProviderStream) Close(ctx context.Context) {
	// no-op
}

func (g *GoogleProvider) convertConversations(conversations []Conversation) (msgWithoutLast []*genai.Content, lastMsg *genai.Content) {
	var chatMessages []*genai.Content
	for _, conversation := range conversations {
		role := ""
		if conversation.Role == ConversationRoleUser {
			role = "user"
		}
		if conversation.Role == ConversationRoleSystem {
			role = "model"
		}
		if role == "" {
			return nil, nil
		}

		chatMessages = append(chatMessages, &genai.Content{
			Parts: []genai.Part{
				genai.Text(conversation.Text),
			},
			Role: role,
		})
	}

	return chatMessages[:len(chatMessages)-1], chatMessages[len(chatMessages)-1]
}
