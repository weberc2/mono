-- This is the Athena/Hive DDL for the blog_analytics table. This was manually
-- invoked because there is no support in CloudFormation for this except for
-- expensive Glue things or custom resources (effort, complexity). Changes to
-- the table should be reflected here.

CREATE EXTERNAL TABLE blog_analytics (
    user_agent string,
    source_ip string,
    time string,
    path string,
    continent_code string,
    continent_name string,
    country_code string,
    country_name string,
    region_code string,
    region_name string,
    city string,
    zip string,
    latitude float,
    longitude float
)
ROW FORMAT SERDE 'org.openx.data.jsonserde.JsonSerDe'
LOCATION 's3://988080168334-prod-blog-analytics-analytics/';


CREATE OR REPLACE VIEW blog_analytics_typed AS
SELECT
    "user_agent",
    "source_ip",
    CAST(from_iso8601_timestamp(time) AS timestamp) as "time",
    "path",
    "continent_code",
    "continent_name",
    "country_code",
    "country_name",
    "region_code",
    "region_name",
    "city",
    "zip",
    "latitude",
    "longitude"
FROM blog_analytics;
