package ui

//go:generate wget -pP assets https://code.jquery.com/jquery-3.6.0.min.js
//go:generate wget -pP assets https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js
//go:generate wget -pP assets https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css

// https://google.github.io/material-design-icons/#setup-method-2-self-hosting
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.ttf
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.eot
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.woff
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.woff2
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/material-icons.css

//go:generate go run ../../vendor/github.com/rakyll/statik/statik.go -f -src assets
