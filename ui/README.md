# ui

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

If the API is not accessible from `http://localhost:8080`, update `.env.development.local` as needed. Use `.env.development` as a guide for it.

Note, things get a bit confusing with the current setup. If using `serve` here, then the server is accessible over HTTP, but it will hit the API at HTTPS.

### Watches source and auto-rebuilds for development
```
npm run watch
```

If the API is not accessible from `http://localhost:8080`, update `.env.development.local` as needed. Use `.env.development` as a guide for it.

This won't use hot-reloading, but will allow you to serve the UI over `HTTPS` using the API's server.

### Compiles and minifies for production
```
npm run build
```

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).
