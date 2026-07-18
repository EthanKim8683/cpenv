import { createClient as createConnectClient } from "@connectrpc/connect";
import { DescService } from "@bufbuild/protobuf";
import { createConnectTransport } from "@connectrpc/connect-web";

export function createClient<T extends DescService>(service: T) {
  const transport = createConnectTransport({
    baseUrl: `http://localhost:${process.env.EXTENSION_PUBLIC_PORT}`,
  });

  return createConnectClient(service, transport);
}
