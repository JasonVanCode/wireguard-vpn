package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"eitec-vpn/internal/server/models"
	"eitec-vpn/internal/shared/utils"
)

// SessionInfo 会话信息
type SessionInfo struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// Session 会话结构
type Session struct {
	ID        string                 `json:"id"`
	UserID    uint                   `json:"user_id"`
	Username  string                 `json:"username"`
	LoginTime time.Time              `json:"login_time"`
	LastSeen  time.Time              `json:"last_seen"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	IsActive  bool                   `json:"is_active"`
	Data      map[string]interface{} `json:"data"`
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions     map[string]*Session
	userSessions map[uint][]string // 用户ID -> 会话ID列表
	mutex        sync.RWMutex
	timeout      time.Duration
}

// NewSessionManager 创建会话管理器
func NewSessionManager(timeout time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:     make(map[string]*Session),
		userSessions: make(map[uint][]string),
		timeout:      timeout,
	}

	// 启动清理过期会话的协程
	go sm.cleanupExpiredSessions()

	return sm
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(user *models.User, ipAddress, userAgent string) (*Session, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("生成会话ID失败: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		LoginTime: now,
		LastSeen:  now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
		Data:      make(map[string]interface{}),
	}

	// 存储会话
	sm.sessions[sessionID] = session

	// 更新用户会话列表
	if sessions, exists := sm.userSessions[user.ID]; exists {
		sm.userSessions[user.ID] = append(sessions, sessionID)
	} else {
		sm.userSessions[user.ID] = []string{sessionID}
	}

	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("会话不存在")
	}

	if !session.IsActive {
		return nil, errors.New("会话已失效")
	}

	// 检查是否过期
	if time.Since(session.LastSeen) > sm.timeout {
		return nil, errors.New("会话已过期")
	}

	return session, nil
}

// UpdateSession 更新会话
func (sm *SessionManager) UpdateSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("会话不存在")
	}

	session.LastSeen = time.Now()
	return nil
}

// SetSessionData 设置会话数据
func (sm *SessionManager) SetSessionData(sessionID, key string, value interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("会话不存在")
	}

	if session.Data == nil {
		session.Data = make(map[string]interface{})
	}
	session.Data[key] = value

	return nil
}

// GetSessionData 获取会话数据
func (sm *SessionManager) GetSessionData(sessionID, key string) (interface{}, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("会话不存在")
	}

	if session.Data == nil {
		return nil, errors.New("数据不存在")
	}

	value, exists := session.Data[key]
	if !exists {
		return nil, errors.New("数据不存在")
	}

	return value, nil
}

// DestroySession 销毁会话
func (sm *SessionManager) DestroySession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("会话不存在")
	}

	// 从用户会话列表中移除
	if userSessions, exists := sm.userSessions[session.UserID]; exists {
		for i, id := range userSessions {
			if id == sessionID {
				sm.userSessions[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
				break
			}
		}

		// 如果用户没有其他会话，删除用户条目
		if len(sm.userSessions[session.UserID]) == 0 {
			delete(sm.userSessions, session.UserID)
		}
	}

	// 删除会话
	delete(sm.sessions, sessionID)
	return nil
}

// DestroyUserSessions 销毁用户的所有会话
func (sm *SessionManager) DestroyUserSessions(userID uint) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionIDs, exists := sm.userSessions[userID]
	if !exists {
		return nil
	}

	// 销毁所有会话
	for _, sessionID := range sessionIDs {
		delete(sm.sessions, sessionID)
	}

	// 删除用户会话列表
	delete(sm.userSessions, userID)
	return nil
}

// GetUserSessions 获取用户的所有会话
func (sm *SessionManager) GetUserSessions(userID uint) ([]*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessionIDs, exists := sm.userSessions[userID]
	if !exists {
		return []*Session{}, nil
	}

	var sessions []*Session
	for _, sessionID := range sessionIDs {
		if session, exists := sm.sessions[sessionID]; exists && session.IsActive {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// GetAllSessions 获取所有活跃会话
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sessions []*Session
	for _, session := range sm.sessions {
		if session.IsActive && time.Since(session.LastSeen) <= sm.timeout {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// IsUserOnline 检查用户是否在线
func (sm *SessionManager) IsUserOnline(userID uint) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessionIDs, exists := sm.userSessions[userID]
	if !exists {
		return false
	}

	// 检查是否有活跃会话
	for _, sessionID := range sessionIDs {
		if session, exists := sm.sessions[sessionID]; exists {
			if session.IsActive && time.Since(session.LastSeen) <= sm.timeout {
				return true
			}
		}
	}

	return false
}

// GetSessionCount 获取会话统计
func (sm *SessionManager) GetSessionCount() (total, active int) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	total = len(sm.sessions)
	for _, session := range sm.sessions {
		if session.IsActive && time.Since(session.LastSeen) <= sm.timeout {
			active++
		}
	}

	return total, active
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()

		var expiredSessions []string
		now := time.Now()

		// 找出过期的会话
		for sessionID, session := range sm.sessions {
			if !session.IsActive || now.Sub(session.LastSeen) > sm.timeout {
				expiredSessions = append(expiredSessions, sessionID)
			}
		}

		// 删除过期会话
		for _, sessionID := range expiredSessions {
			if session, exists := sm.sessions[sessionID]; exists {
				// 从用户会话列表中移除
				if userSessions, exists := sm.userSessions[session.UserID]; exists {
					for i, id := range userSessions {
						if id == sessionID {
							sm.userSessions[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
							break
						}
					}

					// 如果用户没有其他会话，删除用户条目
					if len(sm.userSessions[session.UserID]) == 0 {
						delete(sm.userSessions, session.UserID)
					}
				}

				delete(sm.sessions, sessionID)
			}
		}

		sm.mutex.Unlock()
	}
}

// generateSessionID 生成会话ID
func generateSessionID() (string, error) {
	// 生成32字符的随机字符串作为会话ID
	return utils.GenerateRandomString(32)
}
