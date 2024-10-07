package mm

import (
	"net/url"
)

func Magnet(
	name string,
	infoHash InfoHash,
	trackers ...string,
) string {
	vs := url.Values{}
	for _, tr := range trackers {
		vs.Add("tr", tr)
	}
	if name != "" {
		vs.Add("dn", name)
	}

	// Transmission and Deluge both expect "urn:btih:" to be unescaped. Deluge wants it to be at the
	// start of the magnet link. The InfoHash field is expected to be BitTorrent in this
	// implementation.
	const btihPrefix = "urn:btih:"
	u := url.URL{
		Scheme:   "magnet",
		RawQuery: "xt=" + btihPrefix + infoHash.String(),
	}
	if len(vs) != 0 {
		u.RawQuery += "&" + vs.Encode()
	}
	return u.String()
}
