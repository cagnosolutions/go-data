package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/cagnosolutions/go-data/pkg/web/utils"
)

// var defaultErrTmpl = templates.EmbeddedTemplates.Lookup("error-template.html")

var defaultErrTmpl = template.Must(template.New("").Parse(errTmpl))

func HandleErrors() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(p) > 1 {
			code, err := strconv.Atoi(p[1])
			if err != nil {
				code := http.StatusExpectationFailed
				http.Error(w, http.StatusText(code), code)
				return
			}

			err = defaultErrTmpl.Execute(
				w, struct {
					ErrorCode     int
					ErrorText     string
					ErrorTextLong string
				}{
					ErrorCode:     code,
					ErrorText:     http.StatusText(code),
					ErrorTextLong: utils.HTTPCodesLongFormat[code],
				},
			)
			if err != nil {
				code := http.StatusExpectationFailed
				http.Error(w, http.StatusText(code), code)
				return
			}
		}
	}
	return http.HandlerFunc(fn)
}

var errTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no"/>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <!--[if lt IE 9]>
    <script src="//html5shim.googlecode.com/svn/trunk/html5.js"></script>
    <![endif]-->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossorigin="anonymous">
    <link href='//fonts.googleapis.com/css?family=Lato:100,300,400,700,900,100italic,300italic,400italic,700italic,900italic'
          rel='stylesheet' type='text/css'>
    <title>Ooops, something went wrong!</title>
</head>

<body>

<!-- navigation -->
<div class="container">
    <nav class="navbar fixed-top navbar-expand-lg navbar-light bg-light">
        <div class="container">
            <a class="navbar-brand" href="#">Ooops, something went wrong!</a>
            <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNavAltMarkup" aria-controls="navbarNavAltMarkup" aria-expanded="false" aria-label="Toggle navigation">
                <span class="navbar-toggler-icon"></span>
            </button>
            <div class="collapse navbar-collapse" id="navbarNavAltMarkup">
                <div class="navbar-nav ms-auto">
                    <a class="nav-link" href="/back">Take Me Back!</a>
                </div>
            </div>
        </div>
    </nav>
</div>
<div class="navbar-pad"></div>
<!-- navigation -->

<!-- main section -->
<section class="container">
    <div class="fs-1 fw-light text-center">
        <div class="position-absolute top-50 start-50 translate-middle">
            <span>{{ .ErrorCode }}</span>&nbsp;<span class="text-muted fst-italic">{{ .ErrorText }}</span>
            <br>
            <span class="fs-3 fst-italic fw-lighter text-mutex">
{{ .ErrorTextLong }}
            </span>
            <br>
            <button onclick="history.go(-1)" type="button" class="btn btn-secondary">Please, take me back!</button>
        </div>
    </div>
</section>
<!-- main section -->

<!-- scripts -->
<script src="https://code.jquery.com/jquery-3.6.0.min.js" integrity="sha256-/xUj+3OJU5yExlq6GSYGSHk7tPXikynS7ogEvDej/m4=" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/@popperjs/core@2.10.2/dist/umd/popper.min.js" integrity="sha384-7+zCNj/IqJ95wo16oMtfsKbZ9ccEh31eOz1HGyDuCQ6wgnyJNSYdrPa03rtR1zdB" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.min.js" integrity="sha384-QJHtvGhmr9XOIpI6YVutG+2QOK9T+ZnN4kzFN1RtK3zEFEIsxhlmWl5/YESvpZ13" crossorigin="anonymous"></script>
<script src="https://cdn.jsdelivr.net/npm/just-validate@3.3.1/dist/just-validate.production.min.js" crossorigin="anonymous"></script>
<!-- scripts -->

</body>

<!-- footer -->
<nav class="navbar fixed-bottom navbar-light bg-light">
    <div class="container">
        <a class="navbar-brand" href="#"></a>
        <span class="ms-auto">© Some Company 2021-Present</span>
    </div>
</nav>
<!-- footer -->

</html>`
