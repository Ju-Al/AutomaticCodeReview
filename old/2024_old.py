        scapy.fields.XStrFixedLenField("unencrypted_ephemeral", 0, 32),
        scapy.fields.XStrFixedLenField("encrypted_static", 0, 48),
        scapy.fields.XStrFixedLenField("encrypted_timestamp", 0, 28),
        scapy.fields.XStrFixedLenField("mac1", 0, 16),
        scapy.fields.XStrFixedLenField("mac2", 0, 16),
    ]


class WireguardResponse(Packet):
    name = "Wireguard Response"

    fields_desc = [
        scapy.fields.XLEIntField("sender_index", 0),
        scapy.fields.XLEIntField("receiver_index", 0),
        scapy.fields.XStrFixedLenField("unencrypted_ephemeral", 0, 32),
        scapy.fields.XStrFixedLenField("encrypted_nothing", 0, 16),
        scapy.fields.XStrFixedLenField("mac1", 0, 16),
        scapy.fields.XStrFixedLenField("mac2", 0, 16),
    ]


class WireguardTransport(Packet):
    name = "Wireguard Transport"

    fields_desc = [
        scapy.fields.XLEIntField("receiver_index", 0),
        scapy.fields.XLELongField("counter", 0),
        scapy.fields.XStrField("encrypted_encapsulated_packet", None)
    ]


class WireguardCookieReply(Packet):
    name = "Wireguard Cookie Reply"

    fields_desc = [
        scapy.fields.XLEIntField("receiver_index", 0),
        scapy.fields.XStrFixedLenField("nonce", 0, 24),
        scapy.fields.XStrFixedLenField("encrypted_cookie", 0, 32),
    ]


bind_layers(Wireguard, WireguardInitiation, message_type=1)
bind_layers(Wireguard, WireguardResponse, message_type=2)
bind_layers(Wireguard, WireguardCookieReply, message_type=3)
bind_layers(Wireguard, WireguardTransport, message_type=4)
bind_layers(UDP, Wireguard, dport=51820)
bind_layers(UDP, Wireguard, sport=51820)
