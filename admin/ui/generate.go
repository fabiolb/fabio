package ui

//go:generate wget -pP assets https://code.jquery.com/jquery-3.3.1.min.js
//go:generate wget -pP assets https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/js/materialize.min.js
//go:generate wget -pP assets https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/css/materialize.min.css

// https://google.github.io/material-design-icons/#setup-method-2-self-hosting
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.ttf
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.eot
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.woff
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/MaterialIcons-Regular.woff2
//go:generate wget -nH -nd -pP assets/fonts https://raw.githubusercontent.com/google/material-design-icons/3.0.1/iconfont/material-icons.css

//go:generate go run ../../vendor/github.com/rakyll/statik -f -src assets
