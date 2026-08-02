export * from "./http";

export { authService } from "./auth.service";
export { inboxService } from "./inbox.service";
export { conversationsService } from "./conversations.service";
export { customersService } from "./customers.service";
export { engagementsService } from "./engagements.service";
export { followupsService } from "./followups.service";
export { popsService } from "./pops.service";
export { usersService } from "./users.service";
export { unitsService } from "./units.service";
export { dashboardService } from "./dashboard.service";
export { metaService } from "./meta.service";
export { lossReasonsService } from "./loss-reasons.service";
export { campaignsService } from "./campaigns.service";

export type { LoginResult, SessionResult } from "./auth.service";
export type { ListInboxParams } from "./inbox.service";
export type {
  ListMessagesParams,
  SendMessagePayload,
  UpdatePipelinePayload,
} from "./conversations.service";
export type { ListCustomersParams } from "./customers.service";
export type { ListEngagementsParams } from "./engagements.service";
export type { ListFollowUpsParams } from "./followups.service";
export type { ListPopsParams, PopPayload } from "./pops.service";
export type { CreateUserPayload, UpdateUserPayload } from "./users.service";
export type { DashboardSummaryParams } from "./dashboard.service";
export type { CreateUnitPayload, UpdateUnitPayload } from "@/domain";

import { authService } from "./auth.service";
import { conversationsService } from "./conversations.service";
import { customersService } from "./customers.service";
import { dashboardService } from "./dashboard.service";
import { engagementsService } from "./engagements.service";
import { followupsService } from "./followups.service";
import { inboxService } from "./inbox.service";
import { lossReasonsService } from "./loss-reasons.service";
import { metaService } from "./meta.service";
import { popsService } from "./pops.service";
import { unitsService } from "./units.service";
import { usersService } from "./users.service";
import { campaignsService } from "./campaigns.service";

/** Convenience aggregate. Prefer importing the individual `xService` named exports. */
export const services = {
  auth: authService,
  inbox: inboxService,
  conversations: conversationsService,
  customers: customersService,
  engagements: engagementsService,
  followups: followupsService,
  pops: popsService,
  users: usersService,
  units: unitsService,
  dashboard: dashboardService,
  meta: metaService,
  lossReasons: lossReasonsService,
  campaigns: campaignsService,
};
