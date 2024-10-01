package verification

const (
	lzInstructionsTitle        = "instructions_title"
	lzInstructions             = "instructions"
	lzButtonLabel              = "button_label"
	lzInputLabel               = "input_label"
	lzInputPlaceholder         = "input_placeholder"
	lzExitChannelButtonLabel   = "exit_channel_button_label"
	lzAlreadyVerifiedResp      = "already_verified_resp"
	lzSuccessfullyVerifiedResp = "successfully_verified_resp"
	lzVerificationNotFoundResp = "verification_not_found_resp"
	lzVerificationExpiredResp  = "verification_expired_resp"
	lzUnknownErrorResp         = "unknown_error_resp"
)

var lzRus = map[string]string{
	lzInstructionsTitle:        "Верификация аккаунта BYOND",
	lzInstructions:             "В данном канале вы можете верифицировать свой BYOND аккаунт.\nДля верификации:\n- Зайдите в игру и дождитесь полной загрузки интерфейса\n- В правой верхней части экрана найдите панель с вкладками\n- Перейдите во вкладку \"**ООС**\"\n- Используйте кнопку \"**Discord Verification**\"\n- По нажатию кнопки в чат будет выведено сообщение с вашим кодом верификации, код действует 5 минут\n- Скопируйте код из игрового чата\n- Нажмите кнопку \"**Верифицировать аккаунт**\" под этим сообщением\n- В появившейся форме **вставьте код и нажмите кнопку отправки**\n\nПри возникновении трудностей или ошибок, свяжитесь с модераторами прямо в этом канале.",
	lzButtonLabel:              "Верифицировать аккаунт",
	lzInputLabel:               "Код верификации",
	lzInputPlaceholder:         "Вставьте код верификации из игры",
	lzExitChannelButtonLabel:   "Выйти из канала",
	lzSuccessfullyVerifiedResp: ":white_check_mark: Ваш аккаунт успешно верифицирован.",
	lzAlreadyVerifiedResp:      ":white_check_mark: Ваш аккаунт уже верифицирован.",
	lzVerificationNotFoundResp: "Верификация с указанным кодом не найдена. Проверьте код и попробуйте ещё раз.",
	lzVerificationExpiredResp:  "Ваш код верификации устарел и будет заменён. Получите новый код из игры и используйте его.",
	lzUnknownErrorResp:         ":warning: Во время обработки верификации произошла ошибка. Пожалуйста, уведомите администрацию.",
}
