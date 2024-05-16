package types

import "time"

type Issue struct {
	Type  string `json:"type"`
	Issue struct {
		ID            int    `json:"id"`
		NodeID        string `json:"node_id"`
		URL           string `json:"url"`
		RepositoryURL string `json:"repository_url"`
		LabelsURL     string `json:"labels_url"`
		CommentsURL   string `json:"comments_url"`
		EventsURL     string `json:"events_url"`
		HTMLURL       string `json:"html_url"`
		Number        int    `json:"number"`
		State         string `json:"state"`
		StateReason   string `json:"state_reason"`
		Title         string `json:"title"`
		Body          string `json:"body"`
		User          struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"user"`
		Labels []struct {
			ID      int    `json:"id"`
			NodeID  string `json:"node_id"`
			URL     string `json:"url"`
			Name    string `json:"name"`
			Color   string `json:"color"`
			Default bool   `json:"default"`
		} `json:"labels"`
		Assignees         []any     `json:"assignees"`
		AuthorAssociation string    `json:"author_association"`
		Locked            bool      `json:"locked"`
		Comments          int       `json:"comments"`
		ClosedAt          time.Time `json:"closed_at"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
	} `json:"issue"`
	Events []struct {
		Event  string `json:"event"`
		ID     int    `json:"id"`
		NodeID string `json:"node_id"`
		URL    string `json:"url"`
		Actor  struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"actor"`
		CommitID          any       `json:"commit_id"`
		CommitURL         any       `json:"commit_url"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at,omitempty"`
		AuthorAssociation string    `json:"author_association,omitempty"`
		Body              string    `json:"body,omitempty"`
		User              struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"user,omitempty"`
		HTMLURL  string `json:"html_url,omitempty"`
		IssueURL string `json:"issue_url,omitempty"`
		Source   struct {
			Issue struct {
				ID            int    `json:"id"`
				NodeID        string `json:"node_id"`
				URL           string `json:"url"`
				RepositoryURL string `json:"repository_url"`
				LabelsURL     string `json:"labels_url"`
				CommentsURL   string `json:"comments_url"`
				EventsURL     string `json:"events_url"`
				HTMLURL       string `json:"html_url"`
				Number        int    `json:"number"`
				State         string `json:"state"`
				StateReason   string `json:"state_reason"`
				Title         string `json:"title"`
				Body          string `json:"body"`
				User          struct {
					Login             string `json:"login"`
					ID                int    `json:"id"`
					NodeID            string `json:"node_id"`
					AvatarURL         string `json:"avatar_url"`
					GravatarID        string `json:"gravatar_id"`
					URL               string `json:"url"`
					HTMLURL           string `json:"html_url"`
					FollowersURL      string `json:"followers_url"`
					FollowingURL      string `json:"following_url"`
					GistsURL          string `json:"gists_url"`
					StarredURL        string `json:"starred_url"`
					SubscriptionsURL  string `json:"subscriptions_url"`
					OrganizationsURL  string `json:"organizations_url"`
					ReposURL          string `json:"repos_url"`
					EventsURL         string `json:"events_url"`
					ReceivedEventsURL string `json:"received_events_url"`
					Type              string `json:"type"`
					SiteAdmin         bool   `json:"site_admin"`
				} `json:"user"`
				Labels            []any     `json:"labels"`
				Assignees         []any     `json:"assignees"`
				AuthorAssociation string    `json:"author_association"`
				Locked            bool      `json:"locked"`
				ActiveLockReason  string    `json:"active_lock_reason"`
				Comments          int       `json:"comments"`
				ClosedAt          time.Time `json:"closed_at"`
				CreatedAt         time.Time `json:"created_at"`
				UpdatedAt         time.Time `json:"updated_at"`
			} `json:"issue"`
			Type string `json:"type"`
		} `json:"source,omitempty"`
		Label struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"label,omitempty"`
	} `json:"events"`
}
