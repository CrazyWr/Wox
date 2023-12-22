package plugin

type WoxPreviewType = string
type WoxPreviewScrollPosition = string

const (
	WoxPreviewTypeMarkdown = "markdown"
	WoxPreviewTypeText     = "text"
	WoxPreviewTypeImage    = "image" // when type is image, data should be WoxImage.String()
	WoxPreviewTypeUrl      = "url"
	WoxPreviewTypeRemote   = "remote" // when type is remote, data should be url to load WoxPreview
)

const (
	WoxPreviewScrollPositionBottom = "bottom" // scroll to bottom after preview first show
)

type WoxPreview struct {
	PreviewType       WoxPreviewType
	PreviewData       string
	PreviewProperties map[string]string // key support i18n
	ScrollPosition    WoxPreviewScrollPosition
}
