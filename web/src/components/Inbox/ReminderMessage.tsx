import { create } from "@bufbuild/protobuf";
import { FieldMaskSchema, timestampDate } from "@bufbuild/protobuf/wkt";
import { BellIcon, CheckIcon, TrashIcon, XIcon } from "lucide-react";
import { useState } from "react";
import toast from "react-hot-toast";
import { memoServiceClient, reminderServiceClient, userServiceClient } from "@/connect";
import useAsyncEffect from "@/hooks/useAsyncEffect";
import useNavigateTo from "@/hooks/useNavigateTo";
import { handleError } from "@/lib/error";
import { cn } from "@/lib/utils";
import { Memo } from "@/types/proto/api/v1/memo_service_pb";
import { Reminder } from "@/types/proto/api/v1/reminder_service_pb";
import { UserNotification, UserNotification_Status } from "@/types/proto/api/v1/user_service_pb";
import { useTranslate } from "@/utils/i18n";

interface Props {
  notification: UserNotification;
}

function ReminderMessage({ notification }: Props) {
  const t = useTranslate();
  const navigateTo = useNavigateTo();
  const [reminder, setReminder] = useState<Reminder | undefined>(undefined);
  const [memo, setMemo] = useState<Memo | undefined>(undefined);
  const [initialized, setInitialized] = useState<boolean>(false);
  const [hasError, setHasError] = useState<boolean>(false);

  useAsyncEffect(async () => {
    if (!notification.reminderUid) {
      setHasError(true);
      return;
    }

    try {
      // Get the reminder details
      const reminderData = await reminderServiceClient.getReminder({
        name: `reminders/${notification.reminderUid}`,
      });
      setReminder(reminderData);

      // Get the memo
      if (reminderData.memo) {
        const memoData = await memoServiceClient.getMemo({
          name: reminderData.memo,
        });
        setMemo(memoData);
      }

      setInitialized(true);
    } catch (error) {
      handleError(error, () => {}, {
        context: "Failed to fetch reminder",
        onError: () => setHasError(true),
      });
    }
  }, [notification.reminderUid]);

  const handleNavigateToMemo = async () => {
    if (!memo) {
      return;
    }

    navigateTo(`/${memo.name}`);
    if (notification.status === UserNotification_Status.UNREAD) {
      handleArchiveMessage(true);
    }
  };

  const handleArchiveMessage = async (silence = false) => {
    await userServiceClient.updateUserNotification({
      notification: {
        name: notification.name,
        status: UserNotification_Status.ARCHIVED,
      },
      updateMask: create(FieldMaskSchema, { paths: ["status"] }),
    });
    if (!silence) {
      toast.success(t("message.archived-successfully"));
    }
  };

  const handleDeleteMessage = async () => {
    await userServiceClient.deleteUserNotification({
      name: notification.name,
    });
    toast.success(t("message.deleted-successfully"));
  };

  if (!initialized && !hasError) {
    return (
      <div className="w-full px-5 py-4 border-b border-border/60 last:border-b-0 bg-muted/10 animate-pulse">
        <div className="flex items-start gap-3">
          <div className="w-10 h-10 rounded-full bg-muted/50 shrink-0" />
          <div className="flex-1 space-y-3">
            <div className="h-4 bg-muted/50 rounded-md w-2/5" />
            <div className="h-3 bg-muted/40 rounded-md w-3/4" />
            <div className="h-20 bg-muted/30 rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  if (hasError) {
    return (
      <div className="w-full px-5 py-4 border-b border-border/60 last:border-b-0 bg-destructive/[0.04] group">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-destructive/15 flex items-center justify-center shrink-0 ring-1 ring-destructive/20">
              <XIcon className="w-5 h-5 text-destructive" strokeWidth={2} />
            </div>
            <span className="text-sm text-destructive/80 font-medium">{t("inbox.failed-to-load")}</span>
          </div>
          <button
            onClick={handleDeleteMessage}
            className="p-1.5 hover:bg-destructive/15 rounded-lg transition-all duration-150 opacity-0 group-hover:opacity-100"
            title={t("common.delete")}
          >
            <TrashIcon className="w-4 h-4 text-destructive/70 hover:text-destructive transition-colors" strokeWidth={2} />
          </button>
        </div>
      </div>
    );
  }

  const isUnread = notification.status === UserNotification_Status.UNREAD;
  const remindAt = reminder?.remindAt ? timestampDate(reminder.remindAt) : undefined;

  return (
    <div
      className={cn(
        "w-full px-5 py-4 border-b border-border/60 last:border-b-0 transition-all duration-200 group relative",
        isUnread ? "bg-primary/[0.03] hover:bg-primary/[0.05]" : "hover:bg-muted/30",
      )}
    >
      {/* Unread indicator bar */}
      {isUnread && <div className="absolute left-0 top-0 bottom-0 w-0.5 bg-gradient-to-b from-primary to-primary/60" />}

      <div className="flex items-start gap-3">
        {/* Bell Icon */}
        <div className="relative shrink-0">
          <div
            className={cn(
              "w-10 h-10 rounded-full flex items-center justify-center",
              isUnread ? "bg-primary/15 text-primary" : "bg-muted/50 text-muted-foreground",
            )}
          >
            <BellIcon className="w-5 h-5" strokeWidth={2} />
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header */}
          <div className="flex items-center justify-between gap-3 mb-1">
            <div className="flex items-center gap-1.5 flex-wrap min-w-0">
              <span className="font-semibold text-sm text-foreground/95">{t("reminder.notification.title")}</span>
              <span className="text-xs text-muted-foreground/60">
                {notification.createTime &&
                  timestampDate(notification.createTime)?.toLocaleDateString([], { month: "short", day: "numeric" })}{" "}
                at{" "}
                {notification.createTime &&
                  timestampDate(notification.createTime)?.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
              </span>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              {isUnread ? (
                <button
                  onClick={() => handleArchiveMessage()}
                  className="p-1.5 hover:bg-primary/10 rounded-lg transition-all duration-150 opacity-0 group-hover:opacity-100"
                  title={t("common.archive")}
                >
                  <CheckIcon className="w-4 h-4 text-muted-foreground hover:text-primary transition-colors" strokeWidth={2} />
                </button>
              ) : (
                <button
                  onClick={handleDeleteMessage}
                  className="p-1.5 hover:bg-destructive/10 rounded-lg transition-all duration-150 opacity-0 group-hover:opacity-100"
                  title={t("common.delete")}
                >
                  <TrashIcon className="w-4 h-4 text-muted-foreground hover:text-destructive transition-colors" strokeWidth={2} />
                </button>
              )}
            </div>
          </div>

          {/* Reminder time */}
          {remindAt && (
            <div className="text-xs text-muted-foreground mb-2">
              {t("reminder.remind-at")}:{" "}
              {remindAt.toLocaleString(undefined, {
                dateStyle: "medium",
                timeStyle: "short",
              })}
            </div>
          )}

          {/* Memo Preview */}
          {memo && (
            <div
              onClick={handleNavigateToMemo}
              className="p-2 sm:p-3 rounded-lg bg-gradient-to-br from-primary/[0.06] to-primary/[0.03] hover:from-primary/[0.1] hover:to-primary/[0.06] cursor-pointer border border-primary/30 hover:border-primary/50 transition-all duration-200 group/memo shadow-sm hover:shadow"
            >
              <p className="text-sm text-foreground/90 line-clamp-3">
                {memo.content || <span className="italic text-muted-foreground/50">Empty memo</span>}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default ReminderMessage;
