package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"strconv"
	"strings"

	"github.com/makiuchi-d/gozxing"
	gozxingqr "github.com/makiuchi-d/gozxing/qrcode"
	qrcode "github.com/yeqown/go-qrcode/v2"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 二维码操作。
const (
	// QROpGenerate 生成二维码。
	QROpGenerate = "generate"
	// QROpDecode 识别二维码。
	QROpDecode = "decode"
)

// QROperations 返回受支持的操作（顺序稳定）。
func QROperations() []string {
	return []string{QROpGenerate, QROpDecode}
}

// 纠错级别。
const (
	// QRECLevelL 是 L 级（约 7% 容错）。
	QRECLevelL = "L"
	// QRECLevelM 是 M 级（约 15% 容错）。
	QRECLevelM = "M"
	// QRECLevelQ 是 Q 级（约 25% 容错）。
	QRECLevelQ = "Q"
	// QRECLevelH 是 H 级（约 30% 容错）。
	QRECLevelH = "H"
)

// QRECLevels 返回受支持的纠错级别（顺序稳定）。
func QRECLevels() []string {
	return []string{QRECLevelL, QRECLevelM, QRECLevelQ, QRECLevelH}
}

// 二维码参数边界。
const (
	// maxQRTextBytes 是二维码内容的字节上限。
	//
	// 超出后二维码会变得很大，扫描体验急剧下降，因此直接拒绝。
	maxQRTextBytes = 4096
	// minQRSize 是模块边长的下限（像素）。
	minQRSize = 1
	// maxQRSize 是模块边长的上限（像素）。
	maxQRSize = 40
	// minQRMargin 是静默区模块数的下限。
	minQRMargin = 0
	// maxQRMargin 是静默区模块数的上限。
	maxQRMargin = 16
	// defaultQRMargin 是默认静默区模块数。
	defaultQRMargin = 4
	// defaultQRSize 是默认模块边长（像素）。
	defaultQRSize = 6
	// maxQRPixels 是图片的总像素边长上限。
	//
	// 防止「超大模块 + 大内容」把内存打满。
	maxQRPixels = 4096
	// maxQRImageBytes 是待识别图片的字节上限。
	maxQRImageBytes = 4 << 20
)

// QROptions 是二维码渲染参数。
type QROptions struct {
	// Size 是模块边长（像素）；缺省 6。
	Size *int `json:"size"`
	// Margin 是静默区模块数；缺省 4。
	Margin *int `json:"margin"`
	// ECLevel 是纠错级别 L / M / Q / H；缺省 M。
	ECLevel string `json:"ecLevel"`
	// Fg 是前景色（#RRGGBB）；缺省 #000000。
	Fg string `json:"fg"`
	// Bg 是背景色（#RRGGBB）；缺省 #FFFFFF。
	Bg string `json:"bg"`
}

// QRCodeRequest 是 qr-code 工具的请求。
type QRCodeRequest struct {
	// Operation 是操作：generate / decode。
	Operation string `json:"operation"`
	// Text 是二维码内容；generate 时必填。
	Text string `json:"text"`
	// Options 是渲染参数。
	Options QROptions `json:"options"`
	// ImageBase64 是待识别的图片（base64 编码的 PNG / JPEG）。
	//
	// **不**支持 multipart：图片走 JSON 里的 base64。
	ImageBase64 string `json:"imageBase64"`
}

// QRCodeResult 是 qr-code 工具的响应。
type QRCodeResult struct {
	// SVG 是矢量图（generate）。
	SVG string `json:"svg,omitempty"`
	// PNGBase64 是位图的 base64（generate）。
	PNGBase64 string `json:"pngBase64,omitempty"`
	// Dimension 是二维码的模块边长（generate）。
	Dimension int `json:"dimension,omitempty"`
	// ModuleSize 是模块的像素边长（generate）。
	ModuleSize int `json:"moduleSize,omitempty"`
	// Content 是识别出的内容（decode）。
	Content string `json:"content,omitempty"`
	// Format 是识别出的图片格式（decode）。
	Format string `json:"format,omitempty"`
}

// QRCode 生成或识别二维码。
//
// 生成：把内容编码为二维码矩阵，并渲染成 SVG 与 PNG。
// 识别：解码 base64 图片并读取二维码内容。
func QRCode(ctx context.Context, req *QRCodeRequest) (QRCodeResult, error) {
	if err := req.Validate(); err != nil {
		return QRCodeResult{}, err
	}

	switch req.Operation {
	case QROpGenerate:
		return req.generate()
	default:
		return req.decode()
	}
}

// Validate 做字段级校验。
func (r *QRCodeRequest) Validate() error {
	switch r.Operation {
	case QROpGenerate, QROpDecode:
	default:
		return tools.InvalidField("operation", QROperations(), r.Operation)
	}

	switch r.Options.ECLevel {
	case "", QRECLevelL, QRECLevelM, QRECLevelQ, QRECLevelH:
	default:
		return tools.InvalidField("options.ecLevel", QRECLevels(), r.Options.ECLevel)
	}
	if r.Options.Size != nil && (*r.Options.Size < minQRSize || *r.Options.Size > maxQRSize) {
		return tools.InvalidField("options.size",
			[]int{minQRSize, maxQRSize}, *r.Options.Size)
	}
	if r.Options.Margin != nil && (*r.Options.Margin < minQRMargin || *r.Options.Margin > maxQRMargin) {
		return tools.InvalidField("options.margin",
			[]int{minQRMargin, maxQRMargin}, *r.Options.Margin)
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{"options.fg", r.Options.Fg},
		{"options.bg", r.Options.Bg},
	} {
		if item.value == "" {
			continue
		}
		if _, _, ok := parseHexColor(item.value); !ok {
			return tools.InvalidField(item.name, "#RRGGBB", item.value)
		}
	}

	switch r.Operation {
	case QROpGenerate:
		if r.Text == "" {
			return tools.MissingField("text")
		}
		if len(r.Text) > maxQRTextBytes {
			return toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
				"field":    "text",
				"maxBytes": maxQRTextBytes,
			})
		}
	default:
		if r.ImageBase64 == "" {
			return tools.MissingField("imageBase64")
		}
	}
	return nil
}

// generate 生成二维码。
func (r *QRCodeRequest) generate() (QRCodeResult, error) {
	code, err := qrcode.NewWith(r.Text, r.encodeOption())
	if err != nil {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "text",
			"reason": "content cannot be encoded as QR code",
		})
	}

	matrix, err := qrMatrix(code)
	if err != nil {
		return QRCodeResult{}, err
	}

	size := defaultQRSize
	if r.Options.Size != nil {
		size = *r.Options.Size
	}
	margin := defaultQRMargin
	if r.Options.Margin != nil {
		margin = *r.Options.Margin
	}

	fg, bg, err := r.colors()
	if err != nil {
		return QRCodeResult{}, err
	}

	total := (len(matrix) + 2*margin) * size
	if total > maxQRPixels {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":    "options.size",
			"maxBytes": maxQRPixels,
		})
	}

	svg, err := renderSVG(matrix, margin, size, fg, bg)
	if err != nil {
		return QRCodeResult{}, err
	}

	pngBytes, err := renderPNG(matrix, margin, size, fg, bg)
	if err != nil {
		return QRCodeResult{}, err
	}

	return QRCodeResult{
		SVG:        svg,
		PNGBase64:  base64.StdEncoding.EncodeToString(pngBytes),
		Dimension:  len(matrix) + 2*margin,
		ModuleSize: size,
	}, nil
}

// decode 识别二维码。
func (r *QRCodeRequest) decode() (QRCodeResult, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(r.ImageBase64))
	if err != nil {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeInvalidField, map[string]any{
			"field":  "imageBase64",
			"reason": "not valid base64",
		})
	}
	if len(raw) > maxQRImageBytes {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":    "imageBase64",
			"maxBytes": maxQRImageBytes,
		})
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
			"field":     "imageBase64",
			"supported": []string{"png", "jpeg"},
		})
	}

	bounds := img.Bounds()
	if bounds.Dx() > maxQRPixels || bounds.Dy() > maxQRPixels {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeTooLarge, map[string]any{
			"field":    "imageBase64",
			"maxBytes": maxQRPixels,
		})
	}

	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeInternalError, map[string]any{
			"reason": "cannot build bitmap from image",
		})
	}

	result, err := gozxingqr.NewQRCodeReader().Decode(bitmap, nil)
	if err != nil {
		return QRCodeResult{}, toolerr.Newf(toolerr.CodeUnsupportedFormat, map[string]any{
			"field":  "imageBase64",
			"reason": "no QR code found in image",
		})
	}

	return QRCodeResult{
		Content: result.GetText(),
		Format:  format,
	}, nil
}

// encodeOption 返回编码选项。
//
// go-qrcode 的纠错级别类型是**未导出**的，因此不能在本包声明
// 返回该类型的函数，只能直接构造选项。
func (r *QRCodeRequest) encodeOption() qrcode.EncodeOption {
	switch r.Options.ECLevel {
	case QRECLevelL:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case QRECLevelQ:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case QRECLevelH:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}
}

// colors 返回前景色与背景色。
func (r *QRCodeRequest) colors() (color.RGBA, color.RGBA, error) {
	fg := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	bg := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	if r.Options.Fg != "" {
		parsed, _, ok := parseHexColor(r.Options.Fg)
		if !ok {
			return fg, bg, tools.InvalidField("options.fg", "#RRGGBB", r.Options.Fg)
		}
		fg = parsed
	}
	if r.Options.Bg != "" {
		parsed, _, ok := parseHexColor(r.Options.Bg)
		if !ok {
			return fg, bg, tools.InvalidField("options.bg", "#RRGGBB", r.Options.Bg)
		}
		bg = parsed
	}
	return fg, bg, nil
}

// qrMatrix 取出二维码的模块矩阵。
//
// go-qrcode 不直接暴露矩阵，因此走一次 Writer 把它取出来。
func qrMatrix(code *qrcode.QRCode) ([][]bool, error) {
	capture := &matrixCapture{}
	if err := code.Save(capture); err != nil {
		return nil, toolerr.Newf(toolerr.CodeInternalError, map[string]any{
			"reason": "cannot render QR matrix",
		})
	}
	if len(capture.matrix) == 0 {
		return nil, toolerr.Newf(toolerr.CodeInternalError, map[string]any{
			"reason": "empty QR matrix",
		})
	}
	return capture.matrix, nil
}

// matrixCapture 捕获二维码矩阵。
type matrixCapture struct {
	// matrix 是模块矩阵；true 表示深色模块。
	matrix [][]bool
}

// Write 实现 qrcode.Writer。
func (m *matrixCapture) Write(mat qrcode.Matrix) error {
	m.matrix = mat.Bitmap()
	return nil
}

// Close 实现 qrcode.Writer。
func (m *matrixCapture) Close() error {
	return nil
}

// renderSVG 把矩阵渲染成 SVG。
func renderSVG(matrix [][]bool, margin int, size int, fg, bg color.RGBA) (string, error) {
	dimension := len(matrix) + 2*margin
	total := dimension * size

	var b strings.Builder
	fmt.Fprintf(&b,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" shape-rendering="crispEdges">`,
		total, total, dimension, dimension)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`,
		dimension, dimension, hexColor(bg))

	for y := range matrix {
		for x := range matrix {
			if !matrix[y][x] {
				continue
			}
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="1" height="1" fill="%s"/>`,
				x+margin, y+margin, hexColor(fg))
		}
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}

// renderPNG 把矩阵渲染成 PNG。
func renderPNG(matrix [][]bool, margin int, size int, fg, bg color.RGBA) ([]byte, error) {
	dimension := len(matrix) + 2*margin
	total := dimension * size

	img := image.NewRGBA(image.Rect(0, 0, total, total))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	for y := range matrix {
		for x := range matrix {
			if !matrix[y][x] {
				continue
			}
			rect := image.Rect(
				(x+margin)*size, (y+margin)*size,
				(x+margin+1)*size, (y+margin+1)*size,
			)
			draw.Draw(img, rect, &image.Uniform{C: fg}, image.Point{}, draw.Src)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, toolerr.Newf(toolerr.CodeInternalError, map[string]any{
			"reason": "cannot encode PNG",
		})
	}
	return buf.Bytes(), nil
}

// parseHexColor 解析 #RRGGBB 形式的颜色。
func parseHexColor(value string) (color.RGBA, int, bool) {
	text := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(text) != 6 {
		return color.RGBA{}, 0, false
	}
	n, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		return color.RGBA{}, 0, false
	}
	return color.RGBA{
		R: uint8(n >> 16),
		G: uint8(n >> 8),
		B: uint8(n),
		A: 255,
	}, int(n), true
}

// hexColor 把颜色渲染成 #RRGGBB。
func hexColor(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}
