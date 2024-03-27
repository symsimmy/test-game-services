package errcode

const (
	Succeed             = 0 // 成功
	Config_error        = 1 // 配置错误
	item_no_enough      = 2 // 道具不足
	player_offline      = 3 // 玩家不在线
	no_authority        = 4 // 权限不足
	goldon_not_enough   = 5 // 金豆不足
	diamond_not_enough  = 6 // 钻石不足
	coin_not_enough     = 7 // 能量币不足
	request_param_error = 8 // 请求参数错误

	Invalid_login_message   = 11 // 注册信息格式错误
	Invalid_auth_api        = 12 // 注册接口请求失败
	No_bind                 = 13 // 网关服务器绑定失败
	No_dispatch             = 14 // 游戏服务器绑定失败
	Invalid_getuserinfo_api = 15 // 获取用户信息接口失败
	No_deliver              = 16 // 转发到游戏服务器失败

	Invalid_Apollo_Config     = 17 // 配置信息格式非法
	Invalid_Json_Message      = 18 // json信息格式非法
	Illegitimate_Pb_Message   = 19 // proto消息格式非法
	Invalid_Pb_Message        = 20 // proto消息格式非法
	Multiple_accounts_kickoff = 21 // 账号多地登录踢用户下线
	Game_server_down_kickoff  = 22 // 游戏服务器下线踢用户下线
	Dungeon_agree_failed      = 23 // 用户申请加入副本失败
	Backward_client_version   = 25 // 客户端版本落后
	Advanced_client_version   = 26 // 客户端版本超前

	Bad_Request  = 400
	Unauthorized = 401
	Forbidden    = 403
	Not_Found    = 404

	Internal_Server_Error = 500

	rail_unlock                = 1001  // 尾迹已解锁
	trail_lock                 = 1002  // 尾迹未解锁
	trail_equipped             = 1003  // 尾迹已装备
	trail_no_stage             = 1004  // 不满足进阶条件
	no_in_team                 = 2001  // 玩家不在队伍中
	team_no_exists             = 2002  // 队伍不存在
	team_max                   = 2003  // 队伍已满
	team_transfer_error        = 2004  // 队长不止转移给自己
	team_exists                = 2005  // 已有队伍
	team_no_match              = 2006  // 不在匹配中
	in_team                    = 2007  // 在队伍中
	appiled_to_join            = 2008  // 已申请进队
	template_no_exists         = 3001  // 该模板不存在
	over_limit_time            = 4001  // 超过限购时间
	over_limit_num             = 4002  // 超过限购数量
	friend_exists              = 5001  // 已经是你的好友了
	black_list_exists          = 5002  // 已经在黑名单中了
	friend_max                 = 5003  // 好友已满
	target_friend_max          = 5004  // 对方好友已满
	friend_gift_max            = 5005  // 今日送礼已达上限
	target_in_black_list       = 5006  // 对方在你黑名单中
	in_target_black_list       = 5007  // 你被对方拉黑了
	friend_no_exists           = 5008  // 对方不是你的好友
	black_list_no_exists       = 5009  // 对方不在黑名单中
	black_list_max             = 5010  // 黑名单已满
	repeate_gift               = 5011  // 不可重复送礼
	no_gift                    = 5012  // 对方未给你送礼
	repeate_reward_gift        = 5013  // 已领取好用送礼
	no_their_friend            = 5014  // 你不是对方好友
	select_check_in_game_error = 6001  // 本月已选择签到游戏
	checked_in                 = 6002  // 今日已签到
	check_in_max               = 6003  // 本月签到次数已满
	make_up_check_in_error     = 6004  // 补签次数已满
	make_up_error              = 6005  // 不可补签
	emoj_not_equipped          = 7001  // 表情未装备
	animtion_not_equipped      = 7002  // 动作未装备
	mail_no_exists             = 8001  // 邮件不存在或已过期
	level_not_enough           = 9001  // 等级不足
	reborn_level_max           = 9002  // 已到最大重生等级
	mission_not_exists         = 10001 // 任务不存在
	mission_progress_error     = 10002 // 不满足完成条件
	dungeon_start              = 11001 // 副本已经开始
	dungeon_not_exists         = 11002 // 副本不存在
	voice_frequently           = 12001 // 发言太频繁请稍后再试
	test_test                  = 12002
)
