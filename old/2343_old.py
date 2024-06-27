
from moto.events.responses import EventsHandler
from localstack.constants import DEFAULT_PORT_EVENTS_BACKEND
def start_events(port=None, asynchronous=None, update_listener=None):
    port = port or config.PORT_EVENTS
    backend_port = DEFAULT_PORT_EVENTS_BACKEND
    apply_patches()
    return start_moto_server(
        key='events',
        port=port,
        name='Cloudwatch Events',
        asynchronous=asynchronous,
        backend_port=backend_port,
        update_listener=update_listener
    )
    # Patch _put_targets  #2101 Events put-targets does not respond
    def EventsHandler_put_targets(self):
<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
 "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html>
  <head>
    <title>503 Backend is unhealthy</title>
  </head>
  <body>
    <h1>Error 503 Backend is unhealthy</h1>
    <p>Backend is unhealthy</p>
    <h3>Guru Mediation:</h3>
    <p>Details: cache-sea4431-SEA 1645521629 3669629123</p>
    <hr>
    <p>Varnish cache server</p>
  </body>
</html>
