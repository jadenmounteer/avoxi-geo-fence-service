# GeoIP Database

This directory holds the MaxMind GeoLite2-Country database used for IP-to-country lookups.

## Required File

- **`GeoLite2-Country.mmdb`** – Binary database mapping IP addresses to ISO country codes

## Where to Get It

1. Sign up for a free account at [MaxMind GeoLite2](https://www.maxmind.com/en/geolite2/signup)
2. Generate a [license key](https://www.maxmind.com/en/accounts/current/license-key)
3. Download `GeoLite2-Country.tar.gz` from your [account downloads page](https://www.maxmind.com/en/accounts/current/geoip/downloads)
4. Extract the tarball and copy `GeoLite2-Country.mmdb` into this directory

## Keeping Data Up to Date

MaxMind publishes updated GeoLite2 databases regularly. The GeoLite2 EULA requires databases to be updated within 30 days of a new release. To automate updates:

- Use [geoipupdate](https://dev.maxmind.com/geoip/updating-databases/#using-geoip-update) with your MaxMind license key
- Configure `geoipupdate` to output the `.mmdb` file into this `data/` directory
