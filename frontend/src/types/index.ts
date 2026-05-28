export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string | null;
  provider: string;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthProvidersResponse {
  providers: string[];
  registration_enabled: boolean;
}

export interface Win {
  id: string;
  user_id: string;
  title: string;
  description: string | null;
  occurred_on: string;
  category: string;
  tags: Tag[];
  created_at: string;
  updated_at: string;
}

export interface Tag {
  id: string;
  name: string;
  colour: string;
}

export interface Job {
  id: string;
  user_id: string;
  company: string;
  title: string;
  start_date: string;
  end_date: string | null;
  employment_type: string;
  transition_type: string | null;
  location: string | null;
  remote: boolean;
  responsibilities: string | null;
  notes: string | null;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface Certification {
  id: string;
  user_id: string;
  name: string;
  provider: string;
  status: "planning" | "studying" | "scheduled" | "passed" | "expired";
  earned_date: string | null;
  expiry_date: string | null;
  cost: number | null;
  currency: string;
  credential_url: string | null;
  study_notes: string | null;
  study_progress: number;
  created_at: string;
  updated_at: string;
}

export interface Skill {
  id: string;
  user_id: string;
  name: string;
  category: string | null;
  proficiency: number;
  notes: string | null;
  evidence: SkillEvidence[];
  created_at: string;
  updated_at: string;
}

export interface SkillEvidence {
  id: string;
  skill_id: string;
  evidence_type: "win" | "certification" | "job";
  evidence_id: string;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
}

export interface Badge {
  id: string;
  name: string;
  description: string | null;
  icon: string;
  colour: string;
  tier: "bronze" | "silver" | "gold";
  condition_type: "count" | "streak" | "action";
  condition_config: Record<string, unknown>;
  is_default: boolean;
  earned: boolean;
  awarded_at: string | null;
  created_at: string;
}

export interface GamificationProgress {
  total_points: number;
  level: number;
  level_title: string;
  next_level: number;
  next_level_at: number;
  badges_earned: number;
  current_streak: number;
  longest_streak: number;
}

export interface HeatmapEntry {
  date: string;
  count: number;
}

export interface UserStreak {
  current_streak: number;
  longest_streak: number;
  last_active_on: string | null;
}

export interface Award {
  type: "badge" | "level_up";
  badge?: Badge;
  level?: number;
  level_title?: string;
}

export interface ProfileVisibility {
  jobs: boolean;
  certifications: boolean;
  skills: boolean;
  wins: boolean;
  badges: boolean;
}

export interface ProfileSettings {
  username: string | null;
  bio: string | null;
  profile_visibility: ProfileVisibility;
}

export interface PublicProfile {
  name: string;
  bio: string | null;
  level: number;
  level_title: string;
  badges: Badge[];
  jobs?: Job[];
  certifications?: Certification[];
  skills?: Skill[];
  wins?: Win[];
}
