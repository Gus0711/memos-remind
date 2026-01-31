import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import dayjs from "dayjs";
import { Bell, Calendar, CalendarDays, Clock, Moon, Sun } from "lucide-react";
import { useState } from "react";
import { toast } from "react-hot-toast";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import useLoading from "@/hooks/useLoading";
import { useCreateReminder, useDeleteReminder, useRemindersByMemo } from "@/hooks/useReminderQueries";
import { handleError } from "@/lib/error";
import { cn } from "@/lib/utils";
import { Reminder_Status } from "@/types/proto/api/v1/reminder_service_pb";
import { type Translations, useTranslate } from "@/utils/i18n";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  memoName: string;
}

interface QuickOption {
  labelKey: Translations;
  icon: React.ReactNode;
  getDate: () => Date;
}

const quickOptions: QuickOption[] = [
  {
    labelKey: "reminder.shortcuts.in-1-hour",
    icon: <Clock className="h-4 w-4" />,
    getDate: () => dayjs().add(1, "hour").toDate(),
  },
  {
    labelKey: "reminder.shortcuts.in-3-hours",
    icon: <Clock className="h-4 w-4" />,
    getDate: () => dayjs().add(3, "hours").toDate(),
  },
  {
    labelKey: "reminder.shortcuts.tonight",
    icon: <Moon className="h-4 w-4" />,
    getDate: () => dayjs().hour(20).minute(0).second(0).toDate(),
  },
  {
    labelKey: "reminder.shortcuts.tomorrow-morning",
    icon: <Sun className="h-4 w-4" />,
    getDate: () => dayjs().add(1, "day").hour(9).minute(0).second(0).toDate(),
  },
  {
    labelKey: "reminder.shortcuts.tomorrow-evening",
    icon: <Moon className="h-4 w-4" />,
    getDate: () => dayjs().add(1, "day").hour(18).minute(0).second(0).toDate(),
  },
  {
    labelKey: "reminder.shortcuts.in-1-week",
    icon: <CalendarDays className="h-4 w-4" />,
    getDate: () => dayjs().add(1, "week").hour(9).minute(0).second(0).toDate(),
  },
];

function ReminderDialog({ open, onOpenChange, memoName }: Props) {
  const t = useTranslate();
  const [customDate, setCustomDate] = useState<string>(dayjs().add(1, "hour").format("YYYY-MM-DDTHH:mm"));
  const [showCustom, setShowCustom] = useState(false);
  const requestState = useLoading(false);

  const { data: existingReminders = [], refetch } = useRemindersByMemo(memoName);
  const createReminder = useCreateReminder();
  const deleteReminder = useDeleteReminder();

  // Filter to show only pending reminders
  const pendingReminders = existingReminders.filter((r) => r.status === Reminder_Status.PENDING);

  const handleQuickOption = async (option: QuickOption) => {
    const date = option.getDate();

    // Check if the date is in the future
    if (date <= new Date()) {
      toast.error(t("reminder.messages.must-be-future"));
      return;
    }

    await createReminderWithDate(date);
  };

  const handleCustomSubmit = async () => {
    const date = dayjs(customDate).toDate();

    // Check if the date is in the future
    if (date <= new Date()) {
      toast.error(t("reminder.messages.must-be-future"));
      return;
    }

    await createReminderWithDate(date);
  };

  const createReminderWithDate = async (date: Date) => {
    try {
      requestState.setLoading();
      await createReminder.mutateAsync({
        memo: memoName,
        remindAt: timestampFromDate(date),
      });
      toast.success(t("reminder.messages.created"));
      await refetch();
      requestState.setFinish();
      onOpenChange(false);
    } catch (error: unknown) {
      await handleError(error, toast.error, {
        context: "Create reminder",
        onError: () => requestState.setError(),
      });
    }
  };

  const handleDeleteReminder = async (reminderName: string) => {
    try {
      requestState.setLoading();
      await deleteReminder.mutateAsync(reminderName);
      toast.success(t("reminder.messages.deleted"));
      await refetch();
      requestState.setFinish();
    } catch (error: unknown) {
      await handleError(error, toast.error, {
        context: "Delete reminder",
        onError: () => requestState.setError(),
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5" />
            {t("reminder.remind-me")}
          </DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          {/* Existing reminders */}
          {pendingReminders.length > 0 && (
            <div className="flex flex-col gap-2">
              <Label className="text-muted-foreground text-xs">{t("reminder.upcoming-reminders")}</Label>
              <div className="flex flex-col gap-1">
                {pendingReminders.map((reminder) => (
                  <div key={reminder.name} className="flex items-center justify-between p-2 bg-muted/50 rounded-md text-sm">
                    <span>
                      {reminder.remindAt
                        ? timestampDate(reminder.remindAt).toLocaleString(undefined, {
                            dateStyle: "medium",
                            timeStyle: "short",
                          })
                        : ""}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-destructive hover:text-destructive"
                      onClick={() => handleDeleteReminder(reminder.name)}
                      disabled={requestState.isLoading}
                    >
                      {t("common.delete")}
                    </Button>
                  </div>
                ))}
              </div>
              <Separator className="my-2" />
            </div>
          )}

          {/* Quick options */}
          <div className="grid grid-cols-2 gap-2">
            {quickOptions.map((option) => (
              <Button
                key={option.labelKey}
                variant="outline"
                className="justify-start gap-2 h-10"
                onClick={() => handleQuickOption(option)}
                disabled={requestState.isLoading}
              >
                {option.icon}
                <span className="text-sm">{t(option.labelKey)}</span>
              </Button>
            ))}
          </div>

          {/* Custom date toggle */}
          <Button
            variant="ghost"
            className={cn("justify-start gap-2", showCustom && "bg-muted")}
            onClick={() => setShowCustom(!showCustom)}
          >
            <Calendar className="h-4 w-4" />
            {t("reminder.shortcuts.custom")}
          </Button>

          {/* Custom date picker */}
          {showCustom && (
            <div className="flex flex-col gap-2">
              <input
                type="datetime-local"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                value={customDate}
                onChange={(e) => setCustomDate(e.target.value)}
                min={dayjs().format("YYYY-MM-DDTHH:mm")}
              />
              <Button onClick={handleCustomSubmit} disabled={requestState.isLoading || !customDate}>
                {t("common.save")}
              </Button>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={requestState.isLoading}>
            {t("common.cancel")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default ReminderDialog;
