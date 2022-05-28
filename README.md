# OmadaSiteDns

DNS server for hostname and reverse DNS lookups in a omada site.

## Exposed ports

| Port | Protocoll | Purpose    |
| ------ | ----------- | ------------ |
| 53   | UDP & TCP | DNS server |

## Environment variables

| Environment                               | Type   | Description                                                                                             | Default | Required |
| ------------------------------------------- | -------- | --------------------------------------------------------------------------------------------------------- | --------- | ---------- |
| OSD_VERBOSE                               | Bool   | Enhanced logging for debugging purpose                                                                  | False   |          |
| OSD_SITE_URL                              | String | Controller url (Example: https://192.168.0.2)                                                           |         | ✔       |
| OSD_SITE_SITE                             | String | Site name as seen in the web interface                                                                  | Default | ✔       |
| OSD_SITE_USERNAME                         | String | Login username (it is strongly advised to create a seperate account for each application)               |         | ✔       |
| OSD_SITE_PASSWORD                         | String | Login password                                                                                          |         | ✔       |
| OSD_SITE_SKIPVERIFY                       | Bool   | Skip SSL verification (has to be true if the url is an IP address or self signed certificates are used) | False   |          |
| OSD_SERVER_TTL                            | Time   | TTL of the DNS responses                                                                                | 5m      |          |
| OSD_CRAWLER_INTERVALL                     | Time   | Time between client data fetches                                                                        | 5m      |          |
| OSD_CRAWLER_GATEWAY_INCLUDE               | Bool   | Add the Gateway as client if enabled                                                                    | False   |          |
| OSD_CRAWLER_GATEWAY_PRIMARYNET            | String | If this network is present in the gateway its address is used for dns queries                           |         |          |
| OSD_CRAWLER_CONVERTERS_***n***_REGEX      | String | Regex for hostname conversion (n is the numeric converter id(0,1,...))                                  |         |          |
| OSD_CRAWLER_CONVERTERS_***n***_SUBSTITUTE | String | Substitution for the regex (n is the numeric converter id(0,1,...))                                     |         |          |
