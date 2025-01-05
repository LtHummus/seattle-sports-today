# Is there a Seattle home game today?

Powers https://isthereaseattlehomegametoday.com/ to let the public know if there is a home game being played in Seattle today. Use this to plan for traffic accordingly.

## Teams we look at

| Team                | League     | Venue                |
|---------------------|------------|----------------------|
| Seattle Mariners    | MLB        | T-Mobile Park        |
| Seattle Sounders    | MLS        | Lumen Field          |
| Seattle Kraken      | NHL        | Climate Pledge Arena |
| Seattle Seahawks    | NFL        | Lumen Field          |
| Seattle Storm       | WNBA       | Climate Pledge Arena |
| Seattle Reign       | NWSL       | Lumen Field          |
| UW Huskies Football | NCAA Div I | Husky Stadium        |

## Music we look at

We also query for musical events at the following venues

* Climate Pledge Arena
* T-Mobile Park
* Lumen Field
* WAMU Theater (the theater under Lumen Field)

## Technical Details

tl;dr every day at 3:14 am, an AWS EventBridge event fires which triggers a Lambda function. This function queries a bunch of APIs (mostly ESPN) and figures out if there's a home game for a Seattle team. The HTML for the page is rendered and then uploaded to an S3 bucket. Finally, the CloudFront distribution in front of the bucket has its cache invalidated so we can start serving the new page.

Finally, we integrate with https://ntfy.sh/ so I get a little push notification on my phone every morning to make sure that everything is running. The notification reports how many games were found if everything worked, or an error if things did not.

