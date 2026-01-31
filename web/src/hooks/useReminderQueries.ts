import { create } from "@bufbuild/protobuf";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { reminderServiceClient } from "@/connect";
import type { ListRemindersRequest, Reminder_Status } from "@/types/proto/api/v1/reminder_service_pb";
import { CreateReminderRequestSchema, ReminderSchema, UpdateReminderRequestSchema } from "@/types/proto/api/v1/reminder_service_pb";

// Query keys factory
export const reminderKeys = {
  all: ["reminders"] as const,
  lists: () => [...reminderKeys.all, "list"] as const,
  list: (filters?: Partial<ListRemindersRequest>) => [...reminderKeys.lists(), filters] as const,
  listByMemo: (memoName: string) => [...reminderKeys.lists(), { memo: memoName }] as const,
  details: () => [...reminderKeys.all, "detail"] as const,
  detail: (name: string) => [...reminderKeys.details(), name] as const,
};

// Hook to fetch all reminders for current user
export function useReminders(filters?: { memo?: string; status?: Reminder_Status; pageSize?: number }) {
  return useQuery({
    queryKey: reminderKeys.list(filters),
    queryFn: async () => {
      const { reminders } = await reminderServiceClient.listReminders({
        memo: filters?.memo,
        status: filters?.status,
        pageSize: filters?.pageSize,
      });
      return reminders;
    },
  });
}

// Hook to fetch reminders for a specific memo
export function useRemindersByMemo(memoName: string) {
  return useQuery({
    queryKey: reminderKeys.listByMemo(memoName),
    queryFn: async () => {
      const { reminders } = await reminderServiceClient.listReminders({
        memo: memoName,
      });
      return reminders;
    },
    enabled: !!memoName,
  });
}

// Hook to fetch a single reminder by name
export function useReminder(name: string) {
  return useQuery({
    queryKey: reminderKeys.detail(name),
    queryFn: async () => {
      const reminder = await reminderServiceClient.getReminder({ name });
      return reminder;
    },
    enabled: !!name,
  });
}

// Hook to create a new reminder
export function useCreateReminder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: { memo: string; remindAt?: { seconds?: bigint; nanos?: number } }) => {
      const reminder = await reminderServiceClient.createReminder(
        create(CreateReminderRequestSchema, {
          memo: request.memo,
          remindAt: request.remindAt,
        }),
      );
      return reminder;
    },
    onSuccess: (reminder) => {
      // Invalidate all reminder lists
      queryClient.invalidateQueries({ queryKey: reminderKeys.lists() });
      // Also invalidate the specific memo's reminders
      if (reminder.memo) {
        queryClient.invalidateQueries({ queryKey: reminderKeys.listByMemo(reminder.memo) });
      }
    },
  });
}

// Hook to update a reminder
export function useUpdateReminder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      reminder,
      updateMask,
    }: {
      reminder: {
        name: string;
        memo?: string;
        remindAt?: { seconds?: bigint; nanos?: number };
        status?: Reminder_Status;
      };
      updateMask: string[];
    }) => {
      const updated = await reminderServiceClient.updateReminder(
        create(UpdateReminderRequestSchema, {
          reminder: create(ReminderSchema, reminder),
          updateMask: { paths: updateMask },
        }),
      );
      return updated;
    },
    onSuccess: (reminder) => {
      // Update cache for this specific reminder
      queryClient.setQueryData(reminderKeys.detail(reminder.name), reminder);
      // Invalidate lists
      queryClient.invalidateQueries({ queryKey: reminderKeys.lists() });
    },
  });
}

// Hook to delete a reminder
export function useDeleteReminder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (name: string) => {
      await reminderServiceClient.deleteReminder({ name });
      return name;
    },
    onSuccess: (name) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: reminderKeys.detail(name) });
      // Invalidate lists
      queryClient.invalidateQueries({ queryKey: reminderKeys.lists() });
    },
  });
}

// Hook to dismiss a reminder
export function useDismissReminder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (name: string) => {
      const reminder = await reminderServiceClient.dismissReminder({ name });
      return reminder;
    },
    onSuccess: (reminder) => {
      // Update cache for this specific reminder
      queryClient.setQueryData(reminderKeys.detail(reminder.name), reminder);
      // Invalidate lists
      queryClient.invalidateQueries({ queryKey: reminderKeys.lists() });
    },
  });
}
