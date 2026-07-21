import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { FocusService } from "@/gen/focus/v1/focus_service_pb";
import { SubmitService } from "@/gen/submit/v1/submit_service_pb";

const transport = createConnectTransport({
  baseUrl: `http://localhost:${import.meta.env.WXT_PORT}`,
});

export const focusClient = createClient(FocusService, transport);
export const submitClient = createClient(SubmitService, transport);
