# Photo Server

Personal photo server. Scans a directory tree for photos, stores references to those photos in a DB, and runs a basic web server to present the photos to a browser over HTTP.

**Put an image here**

# Limitations

* Only works with jpg/jpeg files
* The DB is clobbered every time photos are re-indexed
* The UI is very simple (**broken**) right now and the UX can be greatly improved
* If the DB is re-indexed, UUIDs are re-generated so all thumbnails are orphaned. Delete them and re-generate.
* No meaningful authentication or administration interface

# Usage

## Index

```
photo-server index

Index all photos and generate a DB for the server to use.

-photos-directory     /path/to/photos

                      Path to where photos are stored.

-database             /path/to/photos.db

                      Path where the DB should be created.

-thumbnails           true|false

                      Generate thumbnails while indexing.
                      Optional. Defaults to true.

-thumbnails-directory /path/to/thumbnails

                      Path where thumbnails should be stored.
                      Optional unless -thumbnails is true.

-workers              number

                      Number of workers to run concurrently.
                      Optional. Defaults to 1.
```

### Example

```
go run main.go index \
  -photos-directory ~/photo-server-data/FamilyPhotos \
  -database ~/photo-server-data/photos.db \
  -thumbnails-directory ~/photo-server-data/thumbs \
  -workers 4
```

## Thumbnails

```
photo-server thumbnails

Generate thumbnails for every photo in the datasource.

-photos-directory     /path/to/photos

                      Path to where photos are stored.

-database             /path/to/photos.db

                      Path where the DB should be created.

-thumbnails-directory /path/to/thumbnails

                      Path where thumbnails should be stored.
                      Optional unless -thumbnails is true.

-overwrite-existing   true|false

                      Clobber existing thumbnails or skip over them.
                      Optional. Defaults to false.

-workers              number

                      Number of workers to run concurrently.
                      Optional. Defaults to 1.
```

### Example

```
go run main.go thumbnails \
  -photos-directory ~/photo-server-data/FamilyPhotos \
  -database ~/photo-server-data/photos.db \
  -thumbnails-directory ~/photo-server-data/thumbs
```

## Serve

```
photo-server thumbnails

Serve the photos web interface over HTTP.

-photos-directory     /path/to/photos

                      Path to where photos are stored.

-database             /path/to/photos.db

                      Path where the DB should be created.

-thumbnails-directory /path/to/thumbnails

                      Path where thumbnails should be stored.
                      Optional unless -thumbnails is true.

-http-port            number

                      Port number to serve over HTTP.
                      Optional. Defaults to 8080.
                      If -https-port is provided, all HTTP
                      traffic will redirect to the HTTPS port.

-https-port           number

                      Port number to serve over HTTPS.
                      Optional. Not used by default.
                      If -https-port is provided, all HTTP
                      traffic will redirect to the HTTPS port.

-https-cert-file      /path/to/cert.pem

                      Path where HTTPS certificate can be found.
                      Optional unless -https-port is provided.

-https-cert-key       /path/to/key.pem

                      Path where HTTPS certificate key can be found.
                      Optional unless -https-port is provided.

-access-code          string

                      Private/secret code used to prevent the public from
                      viewing photos.
```

### Example

```
go run main.go serve \
  -photos-directory ~/photo-server-data/FamilyPhotos \
  -database ~/photo-server-data/photos.db \
  -thumbnails-directory ~/photo-server-data/thumbs \
  -http-port 8080 \
  -https-port 9090 \
  -https-cert-file ~/photo-server-data/cert.pem \
  -https-cert-key ~/photo-server-data/key.pem \
  -access-code "password"
```

# TLS/HTTPS Certificates

Assuming `certbot` is installed, and port `80` is already configured to redirect to port `8080` for the app, a certificate can be obtained like so.

```
sudo certbot certonly --standalone --http-01-port 8080
```

Locally, this can be run to generate certificates for testing.

```
openssl req -nodes -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365
```

# Thumbnails

Although it is not _required_ to generate thumbnails in advance, it is recommended. Disk space is cheap, processing power on basic devices is expensive.

Thumbnails will be generated on-demand as needed.

[mattes/epeg](https://github.com/mattes/epeg) offers incredibly fast thumbnail generation and auto-orientation as well. There are no Go bindings available though. [koofr/epeg](https://github.com/koofr/epeg) is a fork that has deviated quite a bit from the upstream, but offers a [goepeg](https://github.com/koofr/goepeg) library with bindings. Speed is maintained from upstream `epeg`, but auto-orientation is lost. [gothumb](https://github.com/koofr/gothumb/) is offered for that specific use case.

## Minimal State Management

* TODO The filesystem acts as the source of truth. Can we get away without a DB entirely? Maybe just for indexing? Could it be in-memory? Measure specs...
* TODO Relative paths? So that we can migrate any state to a new dir and it'll work as long as -root is valid?
* TODO Sane URL queries (could be Graph behind the scenes...) for API? /year/month/date? /albums/? /filesystem/?
* TODO Predictable/portable key structure? YYYY-MM-DDTHH:MM:SS.SSS~PATH/NAME?~UUID?CRC/MD5? Is an opaque UUID meaningful? I mean, if the scan is done in seconds is that fine? There IS a chance that ~ will be in the name... not ideal. FindLeft(|) then FindRight(|) since neither date nor checksum should have that char?

SHOULD it be an import process? Copying files to predictable paths?... Duplicating every photo sucks... Can't do that. Well...need to figure out what niche this app serves. Right now, the niche is just "works for me"

Need the DB, indexable by our unique key. So...files NOT the source of truth eh? The DB and scan is... How do we prevent dupes? Because the import is name PER directory, which should never be duplicated. Should that just be the key? Well, remember, we want a meaningful sort key. Have to persist and maintain the derived date info for predictable sorting.

# Thumbnail Performance Analysis

```
./scripts/test-all.sh -d ~/path/to/images -l 100
```

# TODO/Misc Notes

1. Architecture... Separation of responsibilities, proper state handling, clean up the GraphQL resolvers. Consistent use of edges/pagination rather than that wacky initial load call.

Seems like minimizing the observers would be good.

1. When done loading everything for a Photos group, no more need to listen.

1. Scanning all .pending is obviously horrible. Scope it to the on-screen elements at least. photos.$el.querySelector()

Can we still get meaningful performance if we build *all* the skeletons up front? Or at least fake out the height? then fill in as many skeletons as are needed? That's complicated...

Paginate the years list? One pagination call for *everything*? Using after()

Is there an easy way to figure out which _group_ is in the viewport? Scope things that way?

Thumbnail loading is way too slow.

Preventing wacky page resizing on picture load would be great.

1. 2016 February...no data?
1. 2017 February...no data?

/year/month/photos
/folder/"some/path/to/"photos
  * Different UI on this? Explicit expand? Consider that we can't really do the right-hand nav for this...

# Deployment

## Google Domains

On Linux, dynamic DNS like so could be configured to point a custom Google subdomain at your server.

```
# /etc/ddclient.conf
protocol=dyndns2
use=web, web=ipinfo.io/ip
server=domains.google.com
ssl=yes
login='<login>'
password='<password>'
<the domain name>
```
