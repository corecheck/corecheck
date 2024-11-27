import cookie from 'cookie';

export async function handle({ event, resolve }) {
	const session = cookie.parse(event.request.headers.get('cookie') || '');
	event.locals.session = session;

	return await resolve(event);
}
