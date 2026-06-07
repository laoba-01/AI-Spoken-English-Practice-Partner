-- MySQL initialization for AI English Tutor
CREATE TABLE IF NOT EXISTS user_conversations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    scene VARCHAR(20) NOT NULL COMMENT 'daily/business/exam',
    title VARCHAR(100) DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_scene (user_id, scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

CREATE TABLE IF NOT EXISTS user_messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    conversation_id BIGINT NOT NULL,
    role VARCHAR(10) NOT NULL COMMENT 'user/assistant',
    content TEXT NOT NULL,
    audio_url VARCHAR(255) DEFAULT NULL,
    correction TEXT DEFAULT NULL COMMENT '纠错JSON',
    pronunciation_score TINYINT DEFAULT NULL COMMENT '0-100',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';
