# krakend-validation-bypass
This plugin reads a list of routes from krakend.json, for whom no validation checks (JWT, RBAC) required.  checks.  Examples of such URLs are `signup`, `login`, `welcome` , etc.  The plugin populates the request context with a variable `bypass` as true, so the next plugins in the chain can read this value and pass them through.
