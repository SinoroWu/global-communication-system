import axios from 'axios';
import type { 
  Conversation, 
  Message, 
  ChannelConfig, 
  SystemStats, 
  SimulatorPayload,
  ContentType 
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const client = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const apiService = {
  async getConversations(platform?: string): Promise<Conversation[]> {
    const params = platform && platform !== 'all' ? { platform } : {};
    const res = await client.get<Conversation[]>('/conversations', { params });
    return res.data;
  },

  async getConversation(id: string): Promise<Conversation> {
    const res = await client.get<Conversation>(`/conversations/${id}`);
    return res.data;
  },

  async getMessages(conversationId: string): Promise<Message[]> {
    const res = await client.get<Message[]>(`/conversations/${conversationId}/messages`);
    return res.data;
  },

  async sendMessage(
    conversationId: string, 
    content: string, 
    contentType: ContentType = 'text',
    mediaUrl?: string
  ): Promise<Message> {
    const res = await client.post<Message>('/messages/send', {
      conversationId,
      content,
      contentType,
      mediaUrl,
    });
    return res.data;
  },

  async markAsRead(conversationId: string): Promise<{ success: boolean }> {
    const res = await client.post<{ success: boolean }>(`/conversations/${conversationId}/read`);
    return res.data;
  },

  async updateMetadata(
    conversationId: string, 
    notes: string, 
    tags: string[]
  ): Promise<Conversation> {
    const res = await client.put<Conversation>(`/conversations/${conversationId}/metadata`, {
      notes,
      tags,
    });
    return res.data;
  },

  async getChannels(): Promise<ChannelConfig[]> {
    const res = await client.get<ChannelConfig[]>('/channels');
    return res.data;
  },

  async saveChannel(config: ChannelConfig): Promise<{ success: boolean }> {
    const res = await client.post<{ success: boolean }>('/channels', config);
    return res.data;
  },

  async getStats(): Promise<SystemStats> {
    const res = await client.get<SystemStats>('/stats');
    return res.data;
  },

  async triggerSimulator(payload: SimulatorPayload): Promise<{ success: boolean; message: Message }> {
    const res = await client.post<{ success: boolean; message: Message }>('/simulator/trigger', payload);
    return res.data;
  },
};
