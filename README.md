# carbon-embed

Carbon app embedding helper.

# Usage

Download the repository, make a folder called `app` containing your carbon app.

Make sure it has `app.lua` or `init.lua` as entry points.

Then, simply run `make` and it will build it.

Make sure you also have all the dependencies of carbon installed, including  [`go-bindata`](https://github.com/jteeuwen/go-bindata).

You should probably grab the the Makefile and carbon-app.go and put it in your own apps' dir, relocation the actual source to a sub folder, which is due to be included in the binary.

## Note

Since there are no carbon command line options anymore, you have to use specific environment variables to control the setup.

They all should be self explanatory.

- `CARBON_HOST`
- `CARBON_PORT`
- `CARBON_PORTS`
- `CARBON_CERT`
- `CARBON_KEY`
- `CARBON_ENABLE_HTTP`
- `CARBON_ENABLE_HTTPS`
- `CARBON_ENABLE_HTTP2`
- `CARBON_STATES`
- `CARBON_WORKERS`
- `CARBON_ROOT`
- `CARBON_DEBUG`
- `CARBON_LOGGER`
- `CARBON_RECOVERY`

# License
MIT
