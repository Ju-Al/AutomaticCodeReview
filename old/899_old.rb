<IfModule mod_ssl.c>
  SSLProtocol all -SSLv2 -SSLv3
  SSLRandomSeed startup builtin
  SSLRandomSeed startup file:/dev/urandom 512
  SSLRandomSeed connect builtin
  SSLRandomSeed connect file:/dev/urandom 512

  AddType application/x-x509-ca-cert .crt
  AddType application/x-pkcs7-crl    .crl

  SSLPassPhraseDialog  builtin
  SSLSessionCache "shmcb:<%= @session_cache %>"
  SSLSessionCacheTimeout 300
<% if @ssl_compression -%>
  SSLCompression On
<% end -%>
  <% if scope.function_versioncmp([@apache_version, '2.4']) >= 0 -%>
  Mutex <%= @ssl_mutex %>
  <% else -%>
  SSLMutex <%= @ssl_mutex %>
  <% end -%>
  SSLCryptoDevice builtin
  SSLHonorCipherOrder On
  SSLCipherSuite <%= @ssl_cipher %>
  SSLProtocol <%= @ssl_protocol %>
<% if @ssl_options -%>
  SSLOptions <%= @ssl_options.compact.join(' ') %>
<% end -%>
</IfModule>
