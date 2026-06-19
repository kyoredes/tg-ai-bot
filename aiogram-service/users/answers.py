from datetime import datetime, timezone

from users.schemas import ClientModel, SubscriptionModel


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


def _format_ts(ts: int) -> str:
    return datetime.fromtimestamp(ts, tz=timezone.utc).strftime("%d.%m.%Y")


async def get_subscription_info_answer(subscription: SubscriptionModel) -> str:
    lines = [
        "🧐 *Ваш тариф*",
        "━━━━━━━━━━━━━━━━",
        "📦 *Тариф:* Бесплатный",
        f"📅 *Начало:* {_format_ts(subscription.starts_at)}",
        f"📅 *Действует до:* {_format_ts(subscription.expires_at)}",
    ]
    return "\n".join(lines)


CHAT_ERROR_ANSWER = (
    "Извините, произошла ошибка. Пожалуйста, напишите в поддержку."
)
