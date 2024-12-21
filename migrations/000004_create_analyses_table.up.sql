-- 000004_create_analyses_table.up.sql
CREATE TABLE IF NOT EXISTS analyses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    year INT NOT NULL,
    semester INT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    id_group INT,
    id_department INT,
    count_stud INT,
    count5 INT DEFAULT 0,
    count4 INT DEFAULT 0,
    count3 INT DEFAULT 0,
    count2 INT DEFAULT 0,
    id_user INT,
    overall DECIMAL(5,2) GENERATED ALWAYS AS (
        (count5 + count4 + count3 + count2) / count_stud * 100
    ) STORED,
    average DECIMAL(5,2) GENERATED ALWAYS AS (
        (count5 * 5 + count4 * 4 + count3 * 3 + count2 * 2) / 
        (count5 + count4 + count3 + count2)
    ) STORED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (id_group) REFERENCES `groups`(id) ON DELETE SET NULL,
    FOREIGN KEY (id_department) REFERENCES departments(id) ON DELETE SET NULL,
    FOREIGN KEY (id_user) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;