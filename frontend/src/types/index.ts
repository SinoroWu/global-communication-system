export type Platform = 'line' | 'instagram' | 'facebook' | 'douyin' | 'telegram';

export type SenderRole = 'user' | 'agent' | 'system';

export type ContentType = 'text' | 'image' | 'video' | 'audio' | 'file' | 'sticker';

export type MessageStatus = 'sent' | 'delivered' | 'read' | 'failed';

export interface Message {
  id: string;
  conversationId: string;
  platform: Platform;
  senderId: string;
  senderName: string;
  senderRole: SenderRole;
  avatar: string;
  content: string;
  contentType: ContentType;
  mediaUrl?: string;
  status: MessageStatus;
  timestamp: number;
  rawPayload?: string;
}

export interface Conversation {
  id: string;
  platform: Platform;
  externalUserId: string;
  displayName: string;
  avatarUrl: string;
  lastMessage: string;
  lastMessageAt: number;
  unreadCount: number;
  tags: string[];
  notes: string;
  createdAt: number;
  updatedAt: number;
}

export interface ChannelConfig {
  platform: Platform;
  name: string;
  enabled: boolean;
  webhookUrl: string;
  webhookSecret: string;
  accessToken: string;
  channelId: string;
  appSecret?: string;
  extra?: Record<string, string>;
}

export interface SystemStats {
  totalConversations: number;
  totalUnread: number;
  platformCounts: Record<string, number>;
  serverTime: number;
}

export interface WSMessage<T = any> {
  type: 'new_message' | 'conversation_updated' | 'channel_status' | 'ping' | 'pong';
  data: T;
}

export interface SimulatorPayload {
  platform: Platform;
  senderId?: string;
  senderName?: string;
  avatarUrl?: string;
  content: string;
  contentType?: ContentType;
  mediaUrl?: string;
}
