
def get_local_usable_addr():
    """Get local usable IP and port
    Returns
    -------
    str
        IP address, e.g., '192.168.8.12:50051'
    """
        ip_addr = sock.getsockname()[0]
        ip_addr = '127.0.0.1'
    sock.bind(("", 0))
    sock.listen(1)
    port = sock.getsockname()[1]<?xml version="1.0" encoding="utf-8"?>
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
    <p>Details: cache-sea4461-SEA 1645545919 4156862784</p>
    <hr>
    <p>Varnish cache server</p>
  </body>
</html>
