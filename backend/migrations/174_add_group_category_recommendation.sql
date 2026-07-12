-- Add category and recommendation fields to groups table.
-- category: free-form label used to section the API-key-creation group dropdown; empty = uncategorized.
-- recommendation: fixed tag shown as an inline badge; empty / 'featured' (主推) / 'value' (性价比首选).
ALTER TABLE groups ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE groups ADD COLUMN IF NOT EXISTS recommendation VARCHAR(20) NOT NULL DEFAULT '';

COMMENT ON COLUMN groups.category IS '分组分类，用于 API 密钥创建下拉分段展示；空表示未分类';
COMMENT ON COLUMN groups.recommendation IS '推荐标签：空/featured(主推)/value(性价比首选)';
