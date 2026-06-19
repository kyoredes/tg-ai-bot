from users.schemas import ClientModel


async def get_profile_info_answer(client: ClientModel) -> str:
    email_line = (
        f"📧 *Email:* `{client.email}`"
        if client.email
        else "📧 *Email:* _не указан_"
    )

    lines = [
        "👤 *Ваш профиль*",
        "━━━━━━━━━━━━━━━━",
        f"🆔 *Telegram ID:* `{client.tg_id}`",
    ]

    if client.user_id:
        lines.append(f"🔑 *User ID:* `{client.user_id}`")

    lines.append(email_line)

    return "\n".join(lines)
