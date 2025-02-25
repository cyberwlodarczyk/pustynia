export interface Message<T extends string, D> {
  type: T
  data: D
}

export type ClientJoinRequest = Message<'join', { code: string }>

export type ServerJoinReply = Message<'join', { ok: boolean }>

export type ServerAnswerRequest = Message<'answer', { challenge: string }>

export type ClientAnswerReply = Message<'answer', { answer: string }>

export type ServerConfirmationRequest = Message<
  'confirmation',
  { challenge: string; answer: string }
>

export type ClientConfirmationReply = Message<'confirmation', { ok: boolean }>

export type ClientChatMessage = Message<'chat', { iv: string; data: string }>

export type ChatJoinMessage = Message<'join', { user: string }>

export type ChatNewMessage = Message<'message', { user: string; message: string }>

export type ChatLeaveMessage = Message<'leave', { user: string }>

export type ServerMessage = ServerAnswerRequest | ServerConfirmationRequest | ServerJoinReply

export type ClientMessage =
  | ClientAnswerReply
  | ClientChatMessage
  | ClientConfirmationReply
  | ClientJoinRequest

export type ChatMessage = ChatJoinMessage | ChatNewMessage | ChatLeaveMessage
