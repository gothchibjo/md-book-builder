// Package pdf renders a standalone HTML file to PDF with Chrome (headless).
package pdf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// FooterTemplate is written at the bottom of every page.
const FooterTemplate = `<div style="font-size:9px;width:100%;color:#57606a;text-align:center;">стр. <span class="pageNumber"></span> из <span class="totalPages"></span></div>`

// DefaultChromePath returns the most likely Google Chrome binary on this host.
func DefaultChromePath() string {
	if p := os.Getenv("MD_BOOK_BUILDER_CHROME"); p != "" {
		return p
	}
	switch runtime.GOOS {
	case "darwin":
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}
	return "google-chrome"
}

// Params control PDF generation.
type Params struct {
	ChromePath string
	Out        string
}

// Render converts an HTML file into a PDF file.
func Render(htmlFile string, p Params) error {
	chrome := p.ChromePath
	if chrome == "" {
		chrome = DefaultChromePath()
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chrome),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
	)

	actx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if err := os.MkdirAll(filepath.Dir(p.Out), 0o755); err != nil {
		return err
	}

	var pdfBytes []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+htmlFile),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(250*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			data, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithHeaderTemplate(`<div></div>`).
				WithFooterTemplate(FooterTemplate).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(0.55).
				WithMarginBottom(0.6).
				WithMarginLeft(0.55).
				WithMarginRight(0.55).
				WithGenerateDocumentOutline(true).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBytes = data
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("render pdf: %w", err)
	}
	return os.WriteFile(p.Out, pdfBytes, 0o644)
}

// Open reveals the produced PDF in the default viewer.
func Open(path string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return exec.Command("open", path).Start()
}
