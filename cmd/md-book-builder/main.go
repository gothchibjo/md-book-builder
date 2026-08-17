package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gothchibjo/md-book-builder/internal/book"
	"github.com/gothchibjo/md-book-builder/internal/buildinfo"
	"github.com/gothchibjo/md-book-builder/internal/config"
	"github.com/gothchibjo/md-book-builder/internal/pdf"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "md-book-builder",
		Short: "Build a single PDF from a set of cross-referenced markdown files",
		Long: "md-book-builder assembles selected markdown documents from a knowledge\n" +
			"base into one self-contained HTML book and renders it to a PDF with page\n" +
			"numbers and a table of contents. Documents are configured in a YAML file;",
	}

	var out, htmlPath string
	var open bool

	buildCmd := &cobra.Command{
		Use:   "build CONFIG.yaml",
		Short: "Assemble and render the book to a PDF",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}
			if out != "" {
				cfg.Out = out
			}

			b, err := book.New(cfg)
			if err != nil {
				return err
			}
			doc, err := b.Build()
			if err != nil {
				return err
			}

			tmp, err := os.MkdirTemp("", "md-book-builder-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmp)
			htmlFile := filepath.Join(tmp, "book.html")
			if err := os.WriteFile(htmlFile, []byte(doc.HTML), 0o644); err != nil {
				return err
			}

			if htmlPath != "" {
				if err := doc.WriteHTMLFile(htmlPath); err != nil {
					return err
				}
				cmd.Println("HTML:", htmlPath)
			}

			if err := pdf.Render(htmlFile, pdf.Params{ChromePath: cfg.ChromePath, Out: cfg.Out, Locale: cfg.Locale}); err != nil {
				return err
			}
			cmd.Printf("PDF: %s (%d documents)\n", cfg.Out, doc.Stats.Documents)
			if open {
				return pdf.Open(cfg.Out)
			}
			return nil
		},
	}
	buildCmd.Flags().StringVarP(&out, "out", "o", "", "override the output PDF path from the config")
	buildCmd.Flags().StringVar(&htmlPath, "html", "", "also write the assembled HTML to this path")
	buildCmd.Flags().BoolVar(&open, "open", false, "open the produced PDF in the default viewer")

	verifyCmd := &cobra.Command{
		Use:   "verify CONFIG.yaml",
		Short: "Check the book structure without launching Chrome",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}
			b, err := book.New(cfg)
			if err != nil {
				return err
			}
			doc, err := b.Build()
			if err != nil {
				return err
			}
			res := book.Stats{
				Documents:     doc.Stats.Documents,
				FrontTables:   doc.Stats.FrontTables,
				TOCEntries:    doc.Stats.TOCEntries,
				InternalLinks: doc.Stats.InternalLinks,
				Flattened:     doc.Stats.Flattened,
				BrokenAnchors: doc.Stats.BrokenAnchors,
			}
			if htmlPath != "" {
				if err := doc.WriteHTMLFile(htmlPath); err != nil {
					return err
				}
			}
			cmd.Printf("documents: %d\n", res.Documents)
			cmd.Printf("frontmatter tables: %d\n", res.FrontTables)
			cmd.Printf("toc entries: %d\n", res.TOCEntries)
			cmd.Printf("internal links: %d\n", res.InternalLinks)
			cmd.Printf("flattened (out-of-book): %d\n", res.Flattened)
			cmd.Printf("broken anchors: %d\n", res.BrokenAnchors)
			if res.BrokenAnchors > 0 {
				return fmt.Errorf("%d broken internal anchor(s)", res.BrokenAnchors)
			}
			return nil
		},
	}
	verifyCmd.Flags().StringVar(&htmlPath, "html", "", "write the assembled HTML to this path")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), buildinfo.Version)
			return err
		},
	}

	expandCmd := &cobra.Command{
		Use:   "expand CONFIG.yaml",
		Short: "Print the ordered document list as an include block",
		Long: "expand collects the documents exactly like build does (globs,\n" +
			"link-roots, excludes) but instead of rendering prints them as a\n" +
			"ready-to-paste include block for manual ordering.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}
			docs, _, err := book.Collect(cfg)
			if err != nil {
				return err
			}
			var b strings.Builder
			b.WriteString("include:\n")
			for _, d := range docs {
				b.WriteString("  - " + d.RelPath + "\n")
			}
			_, err = cmd.OutOrStdout().Write([]byte(b.String()))
			return err
		},
	}

	root.AddCommand(buildCmd, verifyCmd, versionCmd, expandCmd)
	return root
}
