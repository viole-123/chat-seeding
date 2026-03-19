-- ============================================================================
-- TEMPLATES DATABASE - FULL COLLECTION
-- Execute: Get-Content templates_full.sql | docker exec -i postgres psql -U postgres -d uniscore_seeding
-- ============================================================================

-- Clear existing templates (optional - comment out if you want to keep old ones)
-- DELETE FROM templates;

-- ============================================================================
-- GOAL TEMPLATES
-- ============================================================================

-- GOAL - Generic (NULL persona = match all)
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_generic_01','prematch', 'GOAL', 'vi', NULL, '⚽ GOAL! {{player}} ghi bàn phút {{minute}}! Tỷ số {{score}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_generic_02', 'prematch', 'GOAL', 'vi', NULL, '⚽ Bàn thắng! {{player}} ghi bàn cho {{team}} phút {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_generic_01', 'prematch', 'GOAL', 'en', NULL, '⚽ GOAL! {{player}} scores at minute {{minute}}! Score {{score}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_generic_02', 'prematch', 'GOAL', 'en', NULL, '⚽ Goal for {{team}}! {{player}} finds the net at {{minute}}''', 5, true, '{}', 1709395200, 1709395200);

-- GOAL - Messi Hype Persona
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_messi_01', 'prematch', 'GOAL', 'vi', 'messi_hype', '⚽🔥 GOLLLLLLL!!! {{player}} vừa ghi bàn TUYỆT VỜI phút {{minute}}! Không thể tin được! {{score}}', 10, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_messi_02', 'prematch', 'GOAL', 'vi', 'messi_hype', '⚽💥 SIÊUUUU PHẨMMM!!! {{player}} ghi bàn phút {{minute}}! Quá đỉnh! Tỷ số: {{score}}', 10, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_messi_03', 'prematch', 'GOAL', 'vi', 'messi_hype', '⚽🎯 BÙNG NỔOSSSS! {{player}} vừa ghi bàn ở phút {{minute}}! Cầu thủ này quá hay! {{score}}', 10, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_messi_early', 'prematch', 'GOAL', 'vi', 'messi_hype', '⚽⚡ BÀN THẮNG SỚM! {{player}} mở tỷ số phút {{minute}}! {{team}} dẫn trước! {{score}}', 10, true, '{"minute_max": 15}', 1709395200, 1709395200),
('tpl_goal_vi_messi_late', 'prematch', 'GOAL', 'vi', 'messi_hype', '⚽🔥 BÀN THẮNG PHÚT CUỐI! {{player}} ghi bàn phút {{minute}}! Kịch tính quá! {{score}}', 10, true, '{"minute_min": 80}', 1709395200, 1709395200),
('tpl_goal_en_messi_01', 'prematch', 'GOAL', 'en', 'messi_hype', '⚽🔥 GOALLLLLL!!! {{player}} scores an AMAZING goal at {{minute}}''! Incredible! {{score}}', 10, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_messi_02', 'prematch', 'GOAL', 'en', 'messi_hype', '⚽💥 WHAT A GOAL!!! {{player}} at minute {{minute}}! Absolutely brilliant! Score: {{score}}', 10, true, '{}', 1709395200, 1709395200);

-- GOAL - Ronaldo Analyst Persona  
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_ronaldo_01', 'prematch', 'GOAL', 'vi', 'ronaldo_analyst', '⚽ Goal phút {{minute}} của {{player}} cho {{team}}. Tỷ số hiện tại: {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_ronaldo_02', 'prematch', 'GOAL', 'vi', 'ronaldo_analyst', '⚽ {{player}} ghi bàn ở phút {{minute}}. Đây là bàn thắng quan trọng. {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_ronaldo_03', 'prematch', 'GOAL', 'vi', 'ronaldo_analyst', '⚽ Bàn thắng từ {{player}} phút {{minute}}. {{team}} đang dẫn trước {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_ronaldo_01', 'prematch', 'GOAL', 'en', 'ronaldo_analyst', '⚽ Goal at minute {{minute}} by {{player}} for {{team}}. Current score: {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_ronaldo_02', 'prematch', 'GOAL', 'en', 'ronaldo_analyst', '⚽ {{player}} scores at {{minute}}''. An important goal. Score: {{score}}', 8, true, '{}', 1709395200, 1709395200);

-- GOAL - Passionate Fan Persona
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_passionate_01', 'prematch', 'GOAL', 'vi', 'passionate_fan', '⚽❤️ YESSSS! {{player}} GHI BÀNNNNN! Tôi yêu cầu thủ này quá! Phút {{minute}}! {{score}}', 9, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_passionate_02', 'prematch', 'GOAL', 'vi', 'passionate_fan', '⚽🎉 {{player}} VỪA GHI BÀN! Tôi đã biết mà! Phút {{minute}}! Tỷ số {{score}}', 9, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_passionate_03', 'prematch', 'GOAL', 'vi', 'passionate_fan', '⚽💪 ĐÚNG RỒI! {{player}} làm được rồi! Bàn thắng phút {{minute}}! {{team}} chiến thắng! {{score}}', 9, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_passionate_01', 'prematch', 'GOAL', 'en', 'passionate_fan', '⚽❤️ YESSSS! {{player}} SCORES! I LOVE this player! Minute {{minute}}! {{score}}', 9, true, '{}', 1709395200, 1709395200);

-- GOAL - Funny Memester Persona
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_funny_01', 'prematch', 'GOAL', 'vi', 'funny_memester', '⚽😂 {{player}} vừa ghi bàn! Thủ môn đối phương: "Mình là ai? Mình ở đâu?" Phút {{minute}} haha {{score}}', 7, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_funny_02', 'prematch', 'GOAL', 'vi', 'funny_memester', '⚽🤣 BOOM! {{player}} ghi bàn phút {{minute}}! Lưới đối phương: *chuckles* I''m in danger {{score}}', 7, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_funny_03', 'prematch', 'GOAL', 'vi', 'funny_memester', '⚽💀 {{player}}: "Xin phép tôi ghi bàn thôi" *scores* Phút {{minute}} {{score}}', 7, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_funny_01', 'prematch', 'GOAL', 'en', 'funny_memester', '⚽😂 {{player}} scores! Goalkeeper: "Am I a joke to you?" Minute {{minute}} lol {{score}}', 7, true, '{}', 1709395200, 1709395200);

-- GOAL - Tactical Coach Persona
INSERT INTO templates (id,phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_goal_vi_tactical_01', 'prematch', 'GOAL', 'vi', 'tactical_coach', '⚽📊 Goal từ {{player}} phút {{minute}}. Đúng như chiến thuật đã vạch ra. {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_tactical_02', 'prematch', 'GOAL', 'vi', 'tactical_coach', '⚽🎯 {{player}} hoàn thành xuất sắc di chuyển phút {{minute}}. Bàn thắng chiến thuật. {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_vi_tactical_03', 'prematch', 'GOAL', 'vi', 'tactical_coach', '⚽📈 Bàn thắng của {{player}} ở phút {{minute}} cho thấy sự triển khai tốt. {{score}}', 8, true, '{}', 1709395200, 1709395200),
('tpl_goal_en_tactical_01', 'prematch', 'GOAL', 'en', 'tactical_coach', '⚽📊 Goal by {{player}} at {{minute}}''. Executed as planned. Score: {{score}}', 8, true, '{}', 1709395200, 1709395200);

-- ============================================================================
-- YELLOW CARD TEMPLATES
-- ============================================================================

-- YELLOW CARD - Generic
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_generic_01', 'prematch', 'YELLOW_CARD', 'vi', NULL, '🟨 Thẻ vàng cho {{player}} phút {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_generic_02', 'prematch', 'YELLOW_CARD', 'vi', NULL, '🟨 {{player}} nhận thẻ vàng ở phút {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_generic_01', 'prematch', 'YELLOW_CARD', 'en', NULL, '🟨 Yellow card for {{player}} at minute {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_generic_02', 'prematch', 'YELLOW_CARD', 'en', NULL, '🟨 {{player}} booked at {{minute}}''', 5, true, '{}', 1709395200, 1709395200);

-- YELLOW CARD - Messi Hype
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_messi_01', 'prematch', 'YELLOW_CARD', 'vi', 'messi_hype', '🟨⚠️ Ôi không! {{player}} ăn thẻ vàng phút {{minute}}! Cẩn thận đấy!', 7, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_messi_02', 'prematch', 'YELLOW_CARD', 'vi', 'messi_hype', '🟨😱 {{player}} vừa nhận thẻ vàng phút {{minute}}! Nguy hiểm rồi!', 7, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_messi_01', 'prematch', 'YELLOW_CARD', 'en', 'messi_hype', '🟨⚠️ Oh no! {{player}} gets booked at {{minute}}''! Be careful!', 7, true, '{}', 1709395200, 1709395200);

-- YELLOW CARD - Ronaldo Analyst
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_ronaldo_01', 'prematch', 'YELLOW_CARD', 'vi', 'ronaldo_analyst', '🟨 Thẻ vàng cho {{player}} ở phút {{minute}}. Cần kiểm soát tốt hơn.', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_ronaldo_02', 'prematch', 'YELLOW_CARD', 'vi', 'ronaldo_analyst', '🟨 {{player}} nhận thẻ vàng phút {{minute}}. Một quyết định đúng đắn của trọng tài.', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_ronaldo_01', 'prematch', 'YELLOW_CARD', 'en', 'ronaldo_analyst', '🟨 Yellow card for {{player}} at {{minute}}''. Needs better control.', 6, true, '{}', 1709395200, 1709395200);

-- YELLOW CARD - Passionate Fan
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_passionate_01', 'prematch', 'YELLOW_CARD', 'vi', 'passionate_fan', '🟨😤 Thẻ vàng cho {{player}}?! Phút {{minute}}?! Trọng tài ơi!', 7, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_passionate_02', 'prematch', 'YELLOW_CARD', 'vi', 'passionate_fan', '🟨💢 {{player}} ăn thẻ vàng phút {{minute}}! Không công bằng chút nào!', 7, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_passionate_01', 'prematch', 'YELLOW_CARD', 'en', 'passionate_fan', '🟨😤 Yellow card for {{player}}?! At {{minute}}''?! Come on ref!', 7, true, '{}', 1709395200, 1709395200);

-- YELLOW CARD - Funny Memester
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_funny_01', 'prematch', 'YELLOW_CARD', 'vi', 'funny_memester', '🟨😅 {{player}} ăn thẻ vàng phút {{minute}}! Anh này chơi như đang chơi GTA ấy 😂', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_funny_02', 'prematch', 'YELLOW_CARD', 'vi', 'funny_memester', '🟨🤣 Thẻ vàng cho {{player}} phút {{minute}}! Trọng tài: "Đây là warning cuối cùng đấy" lol', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_funny_01', 'prematch', 'YELLOW_CARD', 'en', 'funny_memester', '🟨😅 {{player}} booked at {{minute}}''! Playing like it''s GTA lol 😂', 6, true, '{}', 1709395200, 1709395200);

-- YELLOW CARD - Tactical Coach
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_yellow_vi_tactical_01', 'prematch', 'YELLOW_CARD', 'vi', 'tactical_coach', '🟨 {{player}} nhận thẻ vàng phút {{minute}}. Cần điều chỉnh lối chơi để tránh rủi ro.', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_vi_tactical_02', 'prematch', 'YELLOW_CARD', 'vi', 'tactical_coach', '🟨 Thẻ vàng cho {{player}} ở phút {{minute}}. Phải thay đổi chiến thuật pressing.', 6, true, '{}', 1709395200, 1709395200),
('tpl_yellow_en_tactical_01', 'prematch', 'YELLOW_CARD', 'en', 'tactical_coach', '🟨 {{player}} booked at {{minute}}''. Need to adjust playstyle to avoid risks.', 6, true, '{}', 1709395200, 1709395200);

-- ============================================================================
-- RED CARD TEMPLATES
-- ============================================================================

-- RED CARD - Generic
INSERT INTO templates (id, phase, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_generic_01', 'prematch', 'RED_CARD', 'vi', NULL, '🟥 THẺ ĐỎ! {{player}} bị truất quyền thi đấu phút {{minute}}!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_generic_02', 'prematch', 'RED_CARD', 'vi', NULL, '🟥 {{player}} nhận thẻ đỏ trực tiếp ở phút {{minute}}!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_generic_01', 'prematch', 'RED_CARD', 'en', NULL, '🟥 RED CARD! {{player}} sent off at minute {{minute}}!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_generic_02', 'prematch', 'RED_CARD', 'en', NULL, '🟥 {{player}} receives a straight red card at {{minute}}''!', 10, true, '{}', 1709395200, 1709395200);

-- RED CARD - Messi Hype
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_messi_01', 'RED_CARD', 'vi', 'messi_hype', '🟥😱 THẺ ĐỎỎỎỎ! {{player}} bị đuổi khỏi sân phút {{minute}}! Không thể tin được!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_messi_02', 'RED_CARD', 'vi', 'messi_hype', '🟥💥 OMG! {{player}} ăn thẻ đỏ phút {{minute}}! Trận đấu đảo lộn rồi!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_messi_01', 'RED_CARD', 'en', 'messi_hype', '🟥😱 RED CARDDD! {{player}} sent off at {{minute}}''! Unbelievable!', 10, true, '{}', 1709395200, 1709395200);

-- RED CARD - Ronaldo Analyst
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_ronaldo_01', 'RED_CARD', 'vi', 'ronaldo_analyst', '🟥 Thẻ đỏ cho {{player}} phút {{minute}}. Quyết định này sẽ thay đổi cục diện trận đấu.', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_ronaldo_02', 'RED_CARD', 'vi', 'ronaldo_analyst', '🟥 {{player}} bị truất quyền ở phút {{minute}}. Đội chỉ còn 10 người.', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_ronaldo_01', 'RED_CARD', 'en', 'ronaldo_analyst', '🟥 Red card for {{player}} at {{minute}}''. This will change the game dynamics.', 10, true, '{}', 1709395200, 1709395200);

-- RED CARD - Passionate Fan
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_passionate_01', 'RED_CARD', 'vi', 'passionate_fan', '🟥😡 THẺ ĐỎ CHO {{player}}?! PHÚT {{minute}}?! ĐÂY LÀ TRẬN CẮM GÌ VẬY!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_passionate_02', 'RED_CARD', 'vi', 'passionate_fan', '🟥💔 Không! {{player}} bị đuổi phút {{minute}}! Cầu thủ yêu thích của tôi!', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_passionate_01', 'RED_CARD', 'en', 'passionate_fan', '🟥😡 RED CARD FOR {{player}}?! AT {{minute}}''?! WHAT IS THIS MATCH!', 10, true, '{}', 1709395200, 1709395200);

-- RED CARD - Funny Memester
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_funny_01', 'RED_CARD', 'vi', 'funny_memester', '🟥😂 {{player}} ăn thẻ đỏ phút {{minute}}! Anh ấy: "Oke I''m out" *leaves* 💀', 9, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_funny_02', 'RED_CARD', 'vi', 'funny_memester', '🟥🤣 THẺ ĐỎ! {{player}} phút {{minute}}! Trọng tài: "You''re fired" lmao', 9, true, '{}', 1709395200, 1709395200),
('tpl_red_en_funny_01', 'RED_CARD', 'en', 'funny_memester', '🟥😂 {{player}} red carded at {{minute}}''! Him: "Ight I''m out" *leaves* 💀', 9, true, '{}', 1709395200, 1709395200);

-- RED CARD - Tactical Coach
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_red_vi_tactical_01', 'RED_CARD', 'vi', 'tactical_coach', '🟥 {{player}} rời sân với thẻ đỏ phút {{minute}}. Cần chuyển sang sơ đồ 4-4-1 phòng ngự.', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_vi_tactical_02', 'RED_CARD', 'vi', 'tactical_coach', '🟥 Thẻ đỏ cho {{player}} ở phút {{minute}}. Phải điều chỉnh chiến thuật ngay lập tức.', 10, true, '{}', 1709395200, 1709395200),
('tpl_red_en_tactical_01', 'RED_CARD', 'en', 'tactical_coach', '🟥 {{player}} sent off at {{minute}}''. Need to switch to 4-4-1 defensive formation.', 10, true, '{}', 1709395200, 1709395200);

-- ============================================================================
-- SUBSTITUTION TEMPLATES
-- ============================================================================

-- SUBSTITUTION - Generic
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_generic_01', 'SUBSTITUTION', 'vi', NULL, '🔄 Thay người: {{in_player}} vào sân thay {{out_player}} phút {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_generic_02', 'SUBSTITUTION', 'vi', NULL, '🔄 {{in_player}} thay thế {{out_player}} ở phút {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_generic_01', 'SUBSTITUTION', 'en', NULL, '🔄 Substitution: {{in_player}} replaces {{out_player}} at minute {{minute}}', 5, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_generic_02', 'SUBSTITUTION', 'en', NULL, '🔄 {{in_player}} comes on for {{out_player}} at {{minute}}''', 5, true, '{}', 1709395200, 1709395200);

-- SUBSTITUTION - Messi Hype
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_messi_01', 'SUBSTITUTION', 'vi', 'messi_hype', '🔄✨ Thay người! {{in_player}} vào sân thay {{out_player}} phút {{minute}}! Hy vọng tạo nên sự khác biệt!', 7, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_messi_02', 'SUBSTITUTION', 'vi', 'messi_hype', '🔄🔥 {{in_player}} xuất hiện! Thay {{out_player}} phút {{minute}}! Chuẩn bị cho màn trình diễn!', 7, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_messi_01', 'SUBSTITUTION', 'en', 'messi_hype', '🔄✨ Substitution! {{in_player}} on for {{out_player}} at {{minute}}''! Hope for a game changer!', 7, true, '{}', 1709395200, 1709395200);

-- SUBSTITUTION - Ronaldo Analyst
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_ronaldo_01', 'SUBSTITUTION', 'vi', 'ronaldo_analyst', '🔄 Thay người chiến thuật: {{in_player}} thay {{out_player}} phút {{minute}}', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_ronaldo_02', 'SUBSTITUTION', 'vi', 'ronaldo_analyst', '🔄 {{in_player}} vào sân thay {{out_player}} ở phút {{minute}}. Một sự điều chỉnh hợp lý.', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_ronaldo_01', 'SUBSTITUTION', 'en', 'ronaldo_analyst', '🔄 Tactical substitution: {{in_player}} replaces {{out_player}} at {{minute}}''', 6, true, '{}', 1709395200, 1709395200);

-- SUBSTITUTION - Passionate Fan
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_passionate_01', 'SUBSTITUTION', 'vi', 'passionate_fan', '🔄👏 Cuối cùng! {{in_player}} vào sân thay {{out_player}} phút {{minute}}! Đúng quyết định!', 7, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_passionate_02', 'SUBSTITUTION', 'vi', 'passionate_fan', '🔄💪 {{in_player}} ra sân rồi! Thay {{out_player}} phút {{minute}}! Hy vọng anh ấy làm nên chuyện!', 7, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_passionate_01', 'SUBSTITUTION', 'en', 'passionate_fan', '🔄👏 Finally! {{in_player}} on for {{out_player}} at {{minute}}''! Right decision!', 7, true, '{}', 1709395200, 1709395200);

-- SUBSTITUTION - Funny Memester
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_funny_01', 'SUBSTITUTION', 'vi', 'funny_memester', '🔄😄 {{in_player}} vào sân! {{out_player}} ra về nghỉ phút {{minute}}! "My job here is done" 😂', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_funny_02', 'SUBSTITUTION', 'vi', 'funny_memester', '🔄🤣 Thay người: {{in_player}} in, {{out_player}} out phút {{minute}}! Tag yourself I''m {{out_player}} lol', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_funny_01', 'SUBSTITUTION', 'en', 'funny_memester', '🔄😄 {{in_player}} on! {{out_player}} off at {{minute}}''! "My job here is done" 😂', 6, true, '{}', 1709395200, 1709395200);

-- SUBSTITUTION - Tactical Coach
INSERT INTO templates (id, event_type, lang, persona_id, text, priority, enabled, conditions, created_at, updated_at) VALUES
('tpl_sub_vi_tactical_01', 'SUBSTITUTION', 'vi', 'tactical_coach', '🔄 Điều chỉnh: {{in_player}} thay {{out_player}} phút {{minute}}. Tăng cường tốc độ biên.', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_vi_tactical_02', 'SUBSTITUTION', 'vi', 'tactical_coach', '🔄 {{in_player}} vào sân thay {{out_player}} ở phút {{minute}}. Thay đổi cấu trúc tấn công.', 6, true, '{}', 1709395200, 1709395200),
('tpl_sub_en_tactical_01', 'SUBSTITUTION', 'en', 'tactical_coach', '🔄 Adjustment: {{in_player}} for {{out_player}} at {{minute}}''. Strengthening wing speed.', 6, true, '{}', 1709395200, 1709395200);

-- ============================================================================
-- VERIFICATION QUERY
-- ============================================================================
SELECT 
    event_type,
    lang,
    COALESCE(persona_id, 'NULL') as persona,
    COUNT(*) as count
FROM templates
WHERE enabled = true
GROUP BY event_type, lang, persona_id
ORDER BY event_type, lang, persona;