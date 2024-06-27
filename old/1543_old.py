import re
import json

from functools import reduce

from streamlink.compat import urlparse, range
from streamlink.exceptions import NoStreamsError
from streamlink.plugin import Plugin
from streamlink.plugin.api import http, validate
from streamlink.stream import HDSStream, HLSStream, HTTPStream, RTMPStream
from streamlink.stream.playlist import FLVPlaylist
from streamlink.stream import HLSStream, HTTPStream

COOKIES = {
    "family_filter": "off",
    "ff": "off"
}
QUALITY_MAP = {
    "ld": "240p",
    "sd": "360p",
    "hq": "480p",
    "hd720": "720p",
    "hd1080": "1080p",
    "custom": "live",
    "auto": "hds",
    "source": "hds"
}
STREAM_INFO_URL = "https://www.dailymotion.com/sequence/full/{0}"
STREAM_INFO_URL = "http://www.dailymotion.com/player/metadata/video/{0}"
USER_INFO_URL = "https://api.dailymotion.com/user/{0}"

_url_re = re.compile(r"""
    http(s)?://(\w+\.)?
    dailymotion.com
    (?:
        (/embed)?/(video|live)
        /(?P<media_id>[^_?/]+)
    |
        /(?P<channel_name>[A-Za-z0-9-_]+)
    )
""", re.VERBOSE)

_media_schema = validate.Schema({
    # Chromecast URL is already available in qualities subdict
    # "stream_chromecast_url": validate.url(),
    "qualities": validate.any({
        validate.text: validate.all([{
            "type": validate.text,
            "url": validate.url()
        }])
    })
})
_live_id_schema = validate.Schema(
    {
        "total": int,
        "list": validate.any(
            [],
            [{"id": validate.text}]
        )
    }
)


class DailyMotion(Plugin):
    @classmethod
    def can_handle_url(cls, url):
        return _url_re.match(url)

    def _get_streams_from_media(self, media_id):
        res = http.get(STREAM_INFO_URL.format(media_id), cookies=COOKIES)
        media = http.json(res, schema=_media_schema)

        for quality, streams in media['qualities'].items():
            for stream in streams:
                if stream['type'] == 'application/x-mpegURL':
                    if quality != 'auto':
                        # Avoid duplicate HLS streams with bitrate selector in the URL query
                        continue
                    for s in HLSStream.parse_variant_playlist(self.session, stream['url']).items():
                        yield s
                elif stream['type'] == 'video/mp4':
                    bitrate = quality + 'p'
                    yield bitrate, HTTPStream(self.session, stream['url'])

    def get_live_id(self, username):
        """Get the livestream videoid from a username.
           https://developer.dailymotion.com/tools/apiexplorer#/user/videos/list
        """
        params = {
            "flags": "live_onair"
        }
        api_user_videos = USER_INFO_URL.format(username) + "/videos"
        try:
            res = http.get(api_user_videos.format(username),
                           params=params)
        except Exception as e:
            self.logger.error("invalid username")
            raise NoStreamsError(self.url)

        data = http.json(res, schema=_live_id_schema)
        if data["total"] > 0:
            media_id = data["list"][0]["id"]
            return media_id
        return False

    def _get_streams(self):
        match = _url_re.match(self.url)
        media_id = match.group("media_id")
        username = match.group("channel_name")

        if not media_id and username:
            media_id = self.get_live_id(username)

        if media_id:
            self.logger.debug("Found media ID: {0}", media_id)
            return self._get_streams_from_media(media_id)


__plugin__ = DailyMotion
