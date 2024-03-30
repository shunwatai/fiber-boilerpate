# web

This package for define routes that response with HTML pages. Those pages may commonly shared with other modules.

## Routes

### Home page
Require login.
```
GET /home
```

### Error page
If failed to authenticate, will redirect to this page.
```
GET /error
```
