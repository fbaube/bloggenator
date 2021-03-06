package generator

import (
	"bytes"
	"fmt"
	"html/template"
	FP "path/filepath"
	"strings"
	"github.com/morningconsult/serrors"
)

// ListingData holds the data for the listing page.
type ListingData struct {
	Title      string
	Date       string
	Short      string
	Link       string
	TimeToRead string
	Tags       []*Tag
}

// ListingGenerator Object
type ListingGenerator struct {
	Config *ListingConfig
}

// ListingConfig holds the configuration for the listing page.
type ListingConfig struct {
	Posts  []*Post
	PageTitle string
	IsIndex   bool
	BaseConfig
}

func (pLC *ListingConfig) String() string {
	return fmt.Sprintf("LstgCfg: %s; \n\t PgTtl<%s> IsIdx?<%t> Posts: %+v",
			pLC.BaseConfig.String(), pLC.PageTitle, pLC.IsIndex, pLC.Posts)
}

// Generate starts the listing generation.
func (g *ListingGenerator) Generate() error {
	shortTmplPath := FP.Join("template", "short.html")
	archiveLinkTmplPath := FP.Join("template", "archiveLink.html")
	posts := g.Config.Posts
	// For the ALL POSTS listing AND for the ARCHIVES
	// listing, this template is the MasterPageTemplate.
	t := g.Config.Template
	destDirPath := g.Config.Dest
	targs := *new(IndexHtmlMasterPageTemplateVariableArguments)
	targs.PageTitle = g.Config.PageTitle
	targs.HtmlTitle = g.Config.PageTitle
	shortTmplRaw, err := getTemplate(shortTmplPath)
	if err != nil {
		return err
	}
	var postBlox []string
	for _, post := range posts {
		meta := post.PropSet
		link := fmt.Sprintf("/%s/", post.DirBase)
		ld := ListingData{
			Title:      meta["title"],
			Date:       meta["date"],
			Short:      meta["short"],
			Link:       link,
			Tags:       createTags(meta["tags"]),
			TimeToRead: calculateTimeToRead(post.CntAsHTML),
		}
		execdPostTmplOutput := bytes.Buffer{}
		if err := shortTmplRaw.Execute(&execdPostTmplOutput, ld); err != nil {
			return serrors.Errorf("error executing template %s: %w", shortTmplPath, err)
		}
		postBlox = append(postBlox, execdPostTmplOutput.String())
	}
	htmlBloxFragment := template.HTML(strings.Join(postBlox, "<br />"))
	if g.Config.IsIndex {
		archiveLinkTmplRaw, err := getTemplate(archiveLinkTmplPath)
		if err != nil {
			return err
		}
		execdArchiveLinkTmplOutput := bytes.Buffer{}
		if err := archiveLinkTmplRaw.Execute(&execdArchiveLinkTmplOutput, nil); err != nil {
			return serrors.Errorf("error executing template %s: %w", archiveLinkTmplPath, err)
		}
		htmlBloxFragment = template.HTML(fmt.Sprintf(
			"%s%s", htmlBloxFragment, template.HTML(execdArchiveLinkTmplOutput.String())))
	}
	targs.HtmlContentFrag = htmlBloxFragment
	// WriteIndexHTML(blogProps SU.PropSet, destDirPath, pageTitle,
	// aMetaDesc string, htmlContentFrag template.HTML, t *template.Template)
	if err := WriteIndexHTML(targs, g.Config.BlogProps, destDirPath, t); err != nil {
		return err
	}
	return nil
}

func calculateTimeToRead(input string) string {
	// an average human reads about 200 wpm
	var secondsPerWord = 60.0 / 200.0
	// multiply with the amount of words
	words := secondsPerWord * float64(len(strings.Split(input, " ")))
	// add 12 seconds for each image
	images := 12.0 * strings.Count(input, "<img")
	result := (words + float64(images)) / 60.0
	if result < 1.0 {
		result = 1.0
	}
	return fmt.Sprintf("%.0fm", result)
}
