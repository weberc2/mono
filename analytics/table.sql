-- This is the Athena/Hive DDL for the blog_analytics table. This was manually
-- invoked because there is no support in CloudFormation for this except for
-- expensive Glue things or custom resources (effort, complexity). Changes to
-- the table should be reflected here.

CREATE EXTERNAL TABLE blog_analytics (
    user_agent string,
    source_ip string,
    time string
)
ROW FORMAT SERDE 'org.openx.data.jsonserde.JsonSerDe'
LOCATION 's3://988080168334-prod-blog-analytics-analytics/';


CREATE OR REPLACE VIEW blog_analytics_typed AS
SELECT
    "user_agent",
    "source_ip",
    CAST(from_iso8601_timestamp(time) AS timestamp) as "time"
FROM blog_analytics;
