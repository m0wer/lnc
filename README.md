# github.com/lnproxy/lnc

A minimalist client library for lnd,
originally developed for github.com/lnproxy/lnproxy.

## Example

See https://github.com/lnproxy/lnproxy-relay/blob/main/cmd/nostr-relay/main.go
for an example of initializing a client with node attestation.

## Permissions

For a signet lnproxy nostr relay using the default node attestation, generate a
least-privilege macaroon with:

```bash
lncli \
  --lnddir=/.lnd \
  --network=signet \
  --rpcserver=127.0.0.1:10009 \
  --tlscertpath=/.lnd/tls.cert \
  --macaroonpath=/.lnd/data/chain/bitcoin/signet/admin.macaroon \
  bakemacaroon \
  --save_to=/.lnd/lnproxy.macaroon \
  uri:/lnrpc.Lightning/DecodePayReq \
  uri:/lnrpc.Lightning/LookupInvoice \
  uri:/invoicesrpc.Invoices/AddHoldInvoice \
  uri:/invoicesrpc.Invoices/SubscribeSingleInvoice \
  uri:/invoicesrpc.Invoices/CancelInvoice \
  uri:/invoicesrpc.Invoices/SettleInvoice \
  uri:/routerrpc.Router/SendPaymentV2 \
  uri:/routerrpc.Router/EstimateRouteFee \
  uri:/chainrpc.ChainKit/GetBestBlock \
  uri:/lnrpc.Lightning/GetInfo \
  uri:/lnrpc.Lightning/SignMessage
```

`GetInfo` and `SignMessage` bind the advertised nostr identity to the Lightning
node identity. They may be omitted only when the relay runs with
`--disable-ln-signing`. Adjust the paths and `--network` for other LND layouts
and Bitcoin networks.
