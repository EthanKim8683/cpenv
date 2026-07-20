import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { ExtensionService } from "@/gen/extension/v1/extension_service_pb";

const transport = createConnectTransport({
  baseUrl: `http://localhost:${import.meta.env.WXT_PORT}`,
});

export const serverClient = createClient(ExtensionService, transport);
