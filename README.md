# maryland-farms

Contains source code described in detail on [my website]*https://meggiel.com/project/updated-md-farms/).

## Scripts

Go scripts to crawl the Maryland Horse Board [portal](https://portal.mda.maryland.gov/stables) and geocode stable data using the Google Maps API.

This is my first Go project and there's some refactoring for consistency that could happen, as well as making the command line work for things other than geocoding.

The scripts rely on a gitignored file (scripts/envvars) which should look like:

```
export GEOCODE_URL=https://maps.googleapis.com/maps/api/geocode/json?address=
export GEOCODE_API_KEY=<your google maps API key here>
```

## Plot

The plot directory contains the end result from the various script phases, a simple index.html, and a small javascript file.
