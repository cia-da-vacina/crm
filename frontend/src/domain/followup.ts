import type { FollowUpStatus, PipelineStage } from "./enums";

export interface FollowUp {
  id: string;
  conversation_id: string;
  customer_id: string;
  customer_name: string;
  customer_phone?: string | null;
  unit_id: string;
  pipeline_stage: PipelineStage;
  due_at: string;
  status: FollowUpStatus;
  note: string;
  created_at: string;
  completed_at?: string | null;
}
