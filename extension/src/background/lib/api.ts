import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { ObservationService } from "../../gen/observation/v1/observation_service_pb";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ContestSchema } from "../../gen/contest/v1/contest_pb";

const transport = createConnectTransport({
  baseUrl: `http://localhost:${process.env.EXTENSION_PUBLIC_PORT}`,
});

export const observationClient = createClient(ObservationService, transport);

create(ContestSchema, {});
