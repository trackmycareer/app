export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string | null;
  location: string | null;
  headline: string | null;
  open_to_work: string;
  provider: string;
  is_admin: boolean;
  mfa_enabled: boolean;
  email_verified: boolean;
  email_verified_at: string | null;
  newsletter_opt_in: boolean;
  newsletter_opt_in_at: string | null;
  is_one_time_supporter: boolean;
  is_subscriber: boolean;
  supporter_since: string | null;
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
  work_mode: string;
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
  name: string;
  username: string | null;
  bio: string | null;
  location: string | null;
  headline: string | null;
  open_to_work: string;
  profile_visibility: ProfileVisibility;
}

export interface PublicProfile {
  name: string;
  bio: string | null;
  location: string | null;
  headline: string | null;
  open_to_work: string;
  avatar_url: string | null;
  is_staff: boolean;
  is_supporter: boolean;
  is_custom_domain?: boolean;
  accent_colour?: string;
  linked_accounts: PublicLinkedAccount[];
  level: number;
  level_title: string;
  badges: Badge[];
  jobs?: Job[];
  certifications?: Certification[];
  skills?: Skill[];
  wins?: Win[];
}

export interface JobPreview {
  company: string;
  title: string;
  start_date: string;
  end_date?: string;
  employment_type?: string;
  location?: string;
  work_mode?: string;
  responsibilities?: string;
  notes?: string;
}

export interface CertPreview {
  name: string;
  provider: string;
  status?: string;
  earned_date?: string;
  expiry_date?: string;
  credential_url?: string;
}

export interface SkillPreview {
  name: string;
  category?: string;
  proficiency?: number;
}

export interface WinPreview {
  title: string;
  description?: string;
  occurred_on?: string;
  category?: string;
}

export interface ImportPreview {
  source: string;
  profile?: { name: string; bio: string };
  jobs: JobPreview[];
  certifications: CertPreview[];
  skills: SkillPreview[];
  wins: WinPreview[];
  warnings: string[];
}

export interface ImportResult {
  jobs_created: number;
  certifications_created: number;
  skills_created: number;
  wins_created: number;
  profile_updated: boolean;
}

export interface CompanyResult {
  name: string;
  source: "user" | "companies_house";
  company_number?: string;
  status?: string;
}

export interface JobTitleResult {
  title: string;
  source: "user" | "common";
}

export interface CertSearchResult {
  name: string;
  provider: string;
  source: "user" | "common";
}

export interface SkillSearchResult {
  name: string;
  category: string;
  source: "user" | "common";
}

export interface LinkedAccount {
  id: string;
  provider: "linkedin" | "github" | "website";
  profile_url: string;
  verified: boolean;
  verify_token?: string;
  verified_at: string | null;
  created_at: string;
}

export interface PublicLinkedAccount {
  provider: "linkedin" | "github" | "website";
  profile_url: string;
  verified: boolean;
  verified_at: string | null;
}

export interface LocationResult {
  label: string;
  source: "user" | "photon";
}

export interface MFAStatus {
  enabled: boolean;
  methods: string[];
  passkey_count: number;
  backup_codes_remaining: number;
}

export interface PasskeyInfo {
  id: string;
  name: string;
  created_at: string;
  last_used_at: string | null;
}

export interface CustomDomain {
  id: string;
  domain: string;
  status: "pending" | "active" | "failed";
  ssl_status: string;
  accent_colour: string;
  cname_target: string;
  created_at: string;
  verified_at: string | null;
}
