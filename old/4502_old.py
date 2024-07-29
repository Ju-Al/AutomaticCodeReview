

# Complex example: Schedule a periodic timer

async def inject_async(flow: http.HTTPFlow):
    msg = "hello from mitmproxy! "
    while flow.server_conn.connected:
        ctx.master.commands.call("inject", [flow], False, msg)
        await asyncio.sleep(1)
        msg = msg[1:] + msg[:1]


def websocket_start(flow: http.HTTPFlow):
    asyncio.create_task(inject_async(flow))
