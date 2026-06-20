package renderhtml

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/events"
)

//go:embed index.gohtml
var templateString string

//go:embed seattle-sports-today.css
var cssString string

var (
	pageTemplate *template.Template
	cssTemplate  template.CSS
)

func init() {
	var err error
	pageTemplate, err = template.New("").Parse(templateString)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse template")
	}

	// we embed the special, seattle sports today CSS directly in to the HTML file. Originally, it was served separately,
	// but I wa tired of hand-maintaining the cache-busting hash and figured this was easier. It's less than 2 kB of CSS
	// so no big deal
	cssTemplate = template.CSS(cssString) //#nosec G203 -- entirely static
}

type templateParams struct {
	Events            []*events.Event
	Tomorrow          []*events.Event
	GeneratedDate     string
	FullGeneratedDate template.HTML
	TomorrowHeading   string
	Style             template.CSS
}

func tomorrowHeader(gamesToday, gamesTomorrow bool) string {
	if gamesToday && gamesTomorrow {
		return "And there's more tomorrow...."
	} else if gamesToday {
		// game today but not tomorrow
		return "But nothing is scheduled tomorrow (yet?)...."
	} else if gamesTomorrow {
		// game tomorrow but not today
		return "But things pick up tomorrow...."
	} else {
		// nothing today, nothing tomorrow
		return "And it's all quiet tomorrow too..."
	}
}

func RenderPage(results *events.EventResults, seattleToday time.Time) ([]byte, error) {
	generatedDateString := seattleToday.Format("Monday Jan _2, 2006")
	buf := bytes.NewBuffer(nil)

	// we have to do things this way because by default the Go HTML templating system will strip out comments. We can force it not
	// to do that by passing this as a template.HTML already so the templating system will plonk it in there no questions asked.

	//#nosec G203 -- We generate this with no involvement from the end user
	generatedTimestamp := template.HTML(fmt.Sprintf("<!-- Generated at: %s -->", seattleToday.Format(time.RFC1123)))

	err := pageTemplate.Execute(buf, &templateParams{
		Events:            results.TodayEvent,
		Tomorrow:          results.TomorrowEvents,
		GeneratedDate:     generatedDateString,
		FullGeneratedDate: generatedTimestamp,
		TomorrowHeading:   tomorrowHeader(len(results.TodayEvent) > 0, len(results.TomorrowEvents) > 0),
		Style:             cssTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("renderPage: could not render: %w", err)
	}

	return buf.Bytes(), nil
}
