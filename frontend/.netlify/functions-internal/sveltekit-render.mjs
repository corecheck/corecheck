import { init } from '../serverless.js';

export const handler = init((() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.png","fonts/jetbrains-mono/jetbrains-mono-v12-latin-600.woff","fonts/jetbrains-mono/jetbrains-mono-v12-latin-600.woff2","fonts/jetbrains-mono/jetbrains-mono-v12-latin-regular.woff","fonts/jetbrains-mono/jetbrains-mono-v12-latin-regular.woff2","fonts/remixicon/remixicon.eot","fonts/remixicon/remixicon.glyph.json","fonts/remixicon/remixicon.svg","fonts/remixicon/remixicon.symbol.svg","fonts/remixicon/remixicon.ttf","fonts/remixicon/remixicon.woff","fonts/remixicon/remixicon.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-600.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-600.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-600italic.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-600italic.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-700.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-700.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-700italic.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-700italic.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-italic.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-italic.woff2","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-regular.woff","fonts/source-sans-pro/source-sans-pro-v18-latin_cyrillic-regular.woff2","icons/png/queue.png","icons/svg/activity.svg","icons/svg/ci.svg","icons/svg/pull-request.svg","icons/svg/stats.svg","images/avatars/avatar0.svg","images/avatars/avatar1.svg","images/avatars/avatar2.svg","images/avatars/avatar3.svg","images/avatars/avatar4.svg","images/avatars/avatar5.svg","images/avatars/avatar6.svg","images/avatars/avatar7.svg","images/avatars/avatar8.svg","images/avatars/avatar9.svg","images/favicon/android-chrome-192x192.png","images/favicon/android-chrome-512x512.png","images/favicon/apple-touch-icon.png","images/favicon/browserconfig.xml","images/favicon/favicon-16x16.png","images/favicon/favicon-32x32.png","images/favicon/favicon.ico","images/favicon/mstile-144x144.png","images/favicon/mstile-150x150.png","images/favicon/mstile-310x150.png","images/favicon/mstile-310x310.png","images/favicon/mstile-70x70.png","images/favicon/safari-pinned-tab.svg","images/favicon/site.webmanifest","images/logo.png","images/oauth2/apple.svg","images/oauth2/discord.svg","images/oauth2/facebook.svg","images/oauth2/gitea.svg","images/oauth2/gitee.svg","images/oauth2/github.svg","images/oauth2/gitlab.svg","images/oauth2/google.svg","images/oauth2/instagram.svg","images/oauth2/kakao.svg","images/oauth2/livechat.svg","images/oauth2/microsoft.svg","images/oauth2/oidc.svg","images/oauth2/spotify.svg","images/oauth2/strava.svg","images/oauth2/twitch.svg","images/oauth2/twitter.svg","images/oauth2/vk.svg","images/oauth2/yandex.svg","robots.txt"]),
	mimeTypes: {".png":"image/png",".woff":"font/woff",".woff2":"font/woff2",".eot":"application/vnd.ms-fontobject",".json":"application/json",".svg":"image/svg+xml",".ttf":"font/ttf",".xml":"application/xml",".ico":"image/vnd.microsoft.icon",".webmanifest":"application/manifest+json",".txt":"text/plain"},
	_: {
		client: {"start":"_app/immutable/entry/start.0aa6f026.js","app":"_app/immutable/entry/app.d2929e6f.js","imports":["_app/immutable/entry/start.0aa6f026.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/singletons.53fe462a.js","_app/immutable/entry/app.d2929e6f.js","_app/immutable/chunks/scheduler.ddef2503.js","_app/immutable/chunks/index.fe14b7b1.js"],"stylesheets":[],"fonts":[]},
		nodes: [
			__memo(() => import('../server/nodes/0.js')),
			__memo(() => import('../server/nodes/1.js')),
			__memo(() => import('../server/nodes/2.js')),
			__memo(() => import('../server/nodes/3.js')),
			__memo(() => import('../server/nodes/4.js')),
			__memo(() => import('../server/nodes/5.js')),
			__memo(() => import('../server/nodes/6.js')),
			__memo(() => import('../server/nodes/7.js'))
		],
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/benchmarks",
				pattern: /^\/benchmarks\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 5 },
				endpoint: null
			},
			{
				id: "/jobs",
				pattern: /^\/jobs\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 6 },
				endpoint: null
			},
			{
				id: "/tests",
				pattern: /^\/tests\/?$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 7 },
				endpoint: null
			},
			{
				id: "/[owner]/[repo]/pulls",
				pattern: /^\/([^/]+?)\/([^/]+?)\/pulls\/?$/,
				params: [{"name":"owner","optional":false,"rest":false,"chained":false},{"name":"repo","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/[owner]/[repo]/pulls/[number]",
				pattern: /^\/([^/]+?)\/([^/]+?)\/pulls\/([^/]+?)\/?$/,
				params: [{"name":"owner","optional":false,"rest":false,"chained":false},{"name":"repo","optional":false,"rest":false,"chained":false},{"name":"number","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			}
		],
		matchers: async () => {
			
			return {  };
		}
	}
}
})());
