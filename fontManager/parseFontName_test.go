package fontManager

import "testing"

func TestParseFontName(t *testing.T) {
	tests := []struct {
		in         string
		wantFamily string
		wantStyle  fontStyle
	}{
		{"arial-bold.ttf", "arial", Bold},
		{"arial-BOLD.TTF", "arial", Bold},
		{"Arial-Italic.ttf", "Arial", Italic},
		{"callibri.ttf", "callibri", Regular},
		{"Times-New-Roman-Italic.ttf", "Times-New-Roman", Italic},
		{"myfont.ttf", "myfont", Regular},
		{"myfont-Bold.ttf", "myfont", Bold},
		{"roboto-condensed-Bold.ttf", "roboto-condensed", Bold},
		{"/absolute/path/with/dashes/awesome-font-italic.ttf", "awesome-font", Italic},
		{"relative/dir/SomeFont-Bold.ttf", "SomeFont", Bold},
		{"SuperFont-BoldItalic.ttf", "SuperFont", BoldItalic},
		{"Lighty-Light.ttf", "Lighty", Light},
		{"SemiSerif-SemiBold.ttf", "SemiSerif", SemiBold},
		{"ExtraThing-ExtraBold.ttf", "ExtraThing", ExtraBold},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			fam, style := parseFontName(tt.in)
			if fam != tt.wantFamily {
				t.Fatalf("parseFontName(%q) family = %q; want %q", tt.in, fam, tt.wantFamily)
			}
			if style != tt.wantStyle {
				t.Fatalf("parseFontName(%q) style = %q; want %q", tt.in, style, tt.wantStyle)
			}
		})
	}
}
