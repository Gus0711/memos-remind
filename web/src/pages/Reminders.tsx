import { timestampDate } from "@bufbuild/protobuf/wkt";
import { sortBy } from "lodash-es";
import { BellIcon, BellOffIcon, CheckCircleIcon, ClockIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { toast } from "react-hot-toast";
import { Link } from "react-router-dom";
import Empty from "@/components/Empty";
import MobileHeader from "@/components/MobileHeader";
import { Button } from "@/components/ui/button";
import useMediaQuery from "@/hooks/useMediaQuery";
import { useDeleteReminder, useDismissReminder, useReminders } from "@/hooks/useReminderQueries";
import { handleError } from "@/lib/error";
import { cn } from "@/lib/utils";
import { Reminder, Reminder_Status } from "@/types/proto/api/v1/reminder_service_pb";
import { useTranslate } from "@/utils/i18n";

type FilterType = "all" | "pending" | "triggered" | "dismissed";

const Reminders = () => {
  const t = useTranslate();
  const md = useMediaQuery("md");
  const [filter, setFilter] = useState<FilterType>("pending");

  // Fetch reminders with React Query
  const { data: fetchedReminders = [], refetch } = useReminders();
  const dismissReminder = useDismissReminder();
  const deleteReminder = useDeleteReminder();

  // Sort by remind_at time
  const allReminders = sortBy(fetchedReminders, (reminder: Reminder) => {
    return (reminder.remindAt ? timestampDate(reminder.remindAt) : undefined)?.getTime() || 0;
  });

  const reminders = allReminders.filter((reminder) => {
    if (filter === "pending") return reminder.status === Reminder_Status.PENDING;
    if (filter === "triggered") return reminder.status === Reminder_Status.TRIGGERED;
    if (filter === "dismissed") return reminder.status === Reminder_Status.DISMISSED;
    return true;
  });

  const pendingCount = allReminders.filter((r) => r.status === Reminder_Status.PENDING).length;
  const triggeredCount = allReminders.filter((r) => r.status === Reminder_Status.TRIGGERED).length;
  const dismissedCount = allReminders.filter((r) => r.status === Reminder_Status.DISMISSED).length;

  const handleDismiss = async (reminderName: string) => {
    try {
      await dismissReminder.mutateAsync(reminderName);
      toast.success(t("reminder.messages.dismissed"));
      await refetch();
    } catch (error: unknown) {
      await handleError(error, toast.error, { context: "Dismiss reminder" });
    }
  };

  const handleDelete = async (reminderName: string) => {
    try {
      await deleteReminder.mutateAsync(reminderName);
      toast.success(t("reminder.messages.deleted"));
      await refetch();
    } catch (error: unknown) {
      await handleError(error, toast.error, { context: "Delete reminder" });
    }
  };

  const getStatusIcon = (status: Reminder_Status) => {
    switch (status) {
      case Reminder_Status.PENDING:
        return <ClockIcon className="w-4 h-4 text-yellow-500" />;
      case Reminder_Status.TRIGGERED:
        return <BellIcon className="w-4 h-4 text-blue-500" />;
      case Reminder_Status.DISMISSED:
        return <CheckCircleIcon className="w-4 h-4 text-muted-foreground" />;
      default:
        return <ClockIcon className="w-4 h-4" />;
    }
  };

  const getStatusLabel = (status: Reminder_Status) => {
    switch (status) {
      case Reminder_Status.PENDING:
        return t("reminder.status.pending");
      case Reminder_Status.TRIGGERED:
        return t("reminder.status.triggered");
      case Reminder_Status.DISMISSED:
        return t("reminder.status.dismissed");
      default:
        return "";
    }
  };

  // Extract memo UID from memo name (format: "memos/{uid}")
  const getMemoPath = (memoName: string) => {
    return `/${memoName}`;
  };

  return (
    <section className="@container w-full max-w-5xl min-h-full flex flex-col justify-start items-center sm:pt-3 md:pt-6 pb-8">
      {!md && <MobileHeader />}
      <div className="w-full px-4 sm:px-6">
        <div className="w-full border border-border flex flex-col justify-start items-start rounded-xl bg-background text-foreground overflow-hidden">
          {/* Header */}
          <div className="w-full px-4 py-4 border-b border-border">
            <div className="flex flex-row justify-between items-center">
              <div className="flex flex-row items-center gap-2">
                <BellIcon className="w-5 h-auto text-muted-foreground" />
                <h1 className="text-xl font-semibold">{t("common.reminders")}</h1>
                {pendingCount > 0 && (
                  <span className="ml-1 px-2 py-0.5 text-xs font-medium rounded-full bg-primary text-primary-foreground">
                    {pendingCount}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Filter Tabs */}
          <div className="w-full px-4 py-2 border-b border-border bg-muted/30">
            <div className="flex flex-row gap-1 flex-wrap">
              <button
                onClick={() => setFilter("all")}
                className={cn(
                  "px-3 py-1.5 text-sm font-medium rounded-md transition-colors",
                  filter === "all"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-background/50",
                )}
              >
                {t("common.all")} ({allReminders.length})
              </button>
              <button
                onClick={() => setFilter("pending")}
                className={cn(
                  "px-3 py-1.5 text-sm font-medium rounded-md transition-colors flex items-center gap-1.5",
                  filter === "pending"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-background/50",
                )}
              >
                <ClockIcon className="w-3.5 h-auto" />
                {t("reminder.status.pending")} ({pendingCount})
              </button>
              <button
                onClick={() => setFilter("triggered")}
                className={cn(
                  "px-3 py-1.5 text-sm font-medium rounded-md transition-colors flex items-center gap-1.5",
                  filter === "triggered"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-background/50",
                )}
              >
                <BellIcon className="w-3.5 h-auto" />
                {t("reminder.status.triggered")} ({triggeredCount})
              </button>
              <button
                onClick={() => setFilter("dismissed")}
                className={cn(
                  "px-3 py-1.5 text-sm font-medium rounded-md transition-colors flex items-center gap-1.5",
                  filter === "dismissed"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground hover:bg-background/50",
                )}
              >
                <CheckCircleIcon className="w-3.5 h-auto" />
                {t("reminder.status.dismissed")} ({dismissedCount})
              </button>
            </div>
          </div>

          {/* Reminders List */}
          <div className="w-full">
            {reminders.length === 0 ? (
              <div className="w-full py-16 flex flex-col justify-center items-center">
                <Empty />
                <p className="mt-4 text-sm text-muted-foreground">{t("reminder.no-reminders")}</p>
              </div>
            ) : (
              <div className="flex flex-col divide-y divide-border">
                {reminders.map((reminder: Reminder) => (
                  <ReminderItem
                    key={reminder.name}
                    reminder={reminder}
                    getMemoPath={getMemoPath}
                    getStatusIcon={getStatusIcon}
                    getStatusLabel={getStatusLabel}
                    onDismiss={handleDismiss}
                    onDelete={handleDelete}
                  />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
};

interface ReminderItemProps {
  reminder: Reminder;
  getMemoPath: (memoName: string) => string;
  getStatusIcon: (status: Reminder_Status) => React.ReactNode;
  getStatusLabel: (status: Reminder_Status) => string;
  onDismiss: (name: string) => Promise<void>;
  onDelete: (name: string) => Promise<void>;
}

const ReminderItem: React.FC<ReminderItemProps> = ({ reminder, getMemoPath, getStatusIcon, getStatusLabel, onDismiss, onDelete }) => {
  const t = useTranslate();
  const remindAt = reminder.remindAt ? timestampDate(reminder.remindAt) : undefined;
  const isPast = remindAt && remindAt < new Date();
  const isPending = reminder.status === Reminder_Status.PENDING;

  return (
    <div className="w-full px-4 py-3 hover:bg-muted/30 transition-colors">
      <div className="flex flex-row items-center justify-between gap-4">
        <div className="flex flex-col gap-1 min-w-0 flex-1">
          <div className="flex items-center gap-2">
            {getStatusIcon(reminder.status)}
            <span className="text-xs text-muted-foreground">{getStatusLabel(reminder.status)}</span>
          </div>
          <div className={cn("text-sm font-medium", isPast && isPending && "text-destructive")}>
            {remindAt?.toLocaleString(undefined, {
              dateStyle: "medium",
              timeStyle: "short",
            })}
            {isPast && isPending && <span className="ml-2 text-xs">({t("reminder.past-reminders")})</span>}
          </div>
          <Link to={getMemoPath(reminder.memo)} className="text-xs text-muted-foreground hover:text-foreground truncate" viewTransition>
            {reminder.memo}
          </Link>
        </div>
        <div className="flex flex-row items-center gap-1 shrink-0">
          {isPending && (
            <Button variant="ghost" size="sm" className="h-8 px-2" onClick={() => onDismiss(reminder.name)}>
              <BellOffIcon className="w-4 h-4" />
              <span className="ml-1 hidden sm:inline">{t("reminder.actions.dismiss")}</span>
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2 text-destructive hover:text-destructive"
            onClick={() => onDelete(reminder.name)}
          >
            <Trash2Icon className="w-4 h-4" />
            <span className="ml-1 hidden sm:inline">{t("common.delete")}</span>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Reminders;
