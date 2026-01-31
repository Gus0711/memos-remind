import { BellIcon, BellOffIcon, BellRingIcon, CheckIcon, Loader2Icon } from "lucide-react";
import { Button } from "@/components/ui/button";
import useWebPush from "@/hooks/useWebPush";
import { useTranslate } from "@/utils/i18n";
import SettingGroup from "./SettingGroup";
import SettingRow from "./SettingRow";

const PushNotificationSection = () => {
  const t = useTranslate();
  const { supported, permission, subscribed, enabled, loading, error, subscribe, unsubscribe } = useWebPush();

  // If push is not supported by browser, don't show the section at all
  if (!supported) {
    return null;
  }

  // If server doesn't have VAPID keys configured, show disabled state
  if (!enabled) {
    return (
      <SettingGroup title={t("setting.push-notifications.title")} showSeparator>
        <SettingRow
          label={t("setting.push-notifications.title")}
          description={t("setting.push-notifications.description")}
          tooltip={t("setting.push-notifications.tooltip")}
        >
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <BellOffIcon className="w-4 h-4" />
            <span>{t("setting.push-notifications.not-configured")}</span>
          </div>
        </SettingRow>
      </SettingGroup>
    );
  }

  const handleToggle = async () => {
    if (subscribed) {
      await unsubscribe();
    } else {
      await subscribe();
    }
  };

  const getStatusText = () => {
    if (loading) return t("common.loading");
    if (error) return error;
    if (permission === "denied") return t("setting.push-notifications.permission-denied");
    if (subscribed) return t("setting.push-notifications.enabled");
    return t("setting.push-notifications.disabled");
  };

  const getStatusIcon = () => {
    if (loading) return <Loader2Icon className="w-4 h-4 animate-spin" />;
    if (subscribed) return <BellRingIcon className="w-4 h-4 text-green-500" />;
    if (permission === "denied") return <BellOffIcon className="w-4 h-4 text-red-500" />;
    return <BellIcon className="w-4 h-4 text-muted-foreground" />;
  };

  return (
    <SettingGroup title={t("setting.push-notifications.title")} showSeparator>
      <SettingRow
        label={t("setting.push-notifications.title")}
        description={t("setting.push-notifications.description")}
        tooltip={t("setting.push-notifications.tooltip")}
      >
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            {getStatusIcon()}
            <span>{getStatusText()}</span>
          </div>
          <Button
            variant={subscribed ? "outline" : "default"}
            size="sm"
            onClick={handleToggle}
            disabled={loading || permission === "denied"}
          >
            {loading ? (
              <Loader2Icon className="w-4 h-4 animate-spin" />
            ) : subscribed ? (
              <>
                <CheckIcon className="w-4 h-4 mr-1" />
                {t("common.disable")}
              </>
            ) : (
              <>
                <BellIcon className="w-4 h-4 mr-1" />
                {t("common.enable")}
              </>
            )}
          </Button>
        </div>
      </SettingRow>
    </SettingGroup>
  );
};

export default PushNotificationSection;
