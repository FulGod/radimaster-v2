export interface User {
  id: number; name: string; email: string; role: string; avatar_url: string;
}
export interface MedicalCase {
  id: number; title: string; slug: string; description: string;
  modality: string; body_part: string; diagnosis: string;
  doctor_id: number; doctor: User; folder_name: string;
  phases: Phase[]; is_public: boolean; view_count: number;
  created_at: string;
}
export interface Phase {
  id: number; case_id: number; name: string; folder_name: string;
  position: number; slice_count: number; slices: Slice[];
}
export interface Slice {
  id: number; phase_id: number; position: number;
  file_name: string; width: number; height: number;
}
export interface Classroom {
  id: number; case_id: number; case: MedicalCase;
  doctor_id: number; doctor: User;
  title: string; code: string; is_active: boolean; created_at: string;
}
export interface WSMessage {
  type: string; user_id?: number; user_name?: string; payload?: any;
}
