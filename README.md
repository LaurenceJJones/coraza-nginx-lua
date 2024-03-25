> [!TIP]
> This was a technical exploration for Coraza and CrowdSec Team. Since it was successful we (CrowdSec) implemented this within our WAF component [read more here](https://doc.crowdsec.net/docs/next/appsec/intro)
> 

# coraza-nginx-lua POC
## Docker compose
Using openresty and lua to send a request to coraza "api" that reassembles the request.

## docker compose up
open `http://localhost` you will see DVWA
