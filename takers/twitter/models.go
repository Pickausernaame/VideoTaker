package twitter

type gt struct {
	GuestToken string `json:"guest_token"`
}

type track struct {
	Track struct {
		PlaybackURL string `json:"playbackUrl"`
	} `json:"track"`
}
