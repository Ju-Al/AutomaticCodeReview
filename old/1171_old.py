from pathod import log
import StringIO
from netlib.exceptions import TcpDisconnect
import netlib.tcp

from six.moves import cStringIO


class DummyIO(cStringIO):

    def start_log(self, *args, **kwargs):
        pass

    def get_log(self, *args, **kwargs):
        return ""


def test_disconnect():
    outf = DummyIO()
    rw = DummyIO()
    l = log.ConnectionLogger(outf, False, rw, rw)
    try:
        with l.ctx() as lg:
            lg("Test")
    except TcpDisconnect:
        pass
    assert "Test" in outf.getvalue()