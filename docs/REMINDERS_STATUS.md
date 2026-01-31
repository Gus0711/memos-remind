# Memos Reminders Feature - Status Report

## Complété et fonctionnel

### Backend

- Table `reminder` (SQLite/MySQL/PostgreSQL)
- Table `push_subscription` (SQLite/MySQL/PostgreSQL)
- API gRPC ReminderService (Create, List, Get, Update, Delete, Dismiss)
- API gRPC pour VAPID keys et push subscriptions
- Auto-génération VAPID keys au démarrage
- Runner qui vérifie les reminders dus chaque minute
- Provider Inbox (notifications dans l'app)

### Frontend

- ReminderDialog avec raccourcis temporels
- ReminderButton dans le menu des memos
- Page /reminders avec filtres (Pending/Triggered/Dismissed)
- Badge count dans la sidebar
- Section Push Notifications dans Settings
- Service Worker (sw.js)
- Hook useWebPush pour subscription

## A debugger

### Web Push ne s'envoie pas

- La subscription est créée (log `CreateUserPushSubscription` OK)
- Le reminder est traité (log `processing due reminders count=1`)
- MAIS pas de log d'envoi Web Push
- Hypothèse : le webpush provider n'est pas appelé ou erreur silencieuse

### Actions à faire

1. Ajouter des logs dans `plugin/notification/webpush_provider.go` méthode `Send()`
2. Vérifier que le webpush provider est bien enregistré et appelé dans le reminder runner
3. Vérifier que les subscriptions sont récupérées pour l'utilisateur
4. Tester l'envoi Web Push manuellement

## Fichiers clés

### Backend

| Fichier | Description |
|---------|-------------|
| `server/server.go` | Init VAPID keys |
| `server/runner/reminder/runner.go` | Traitement des reminders |
| `plugin/notification/notification.go` | Service de notification |
| `plugin/notification/webpush_provider.go` | Envoi Web Push |
| `store/push_subscription.go` | CRUD subscriptions |

### Frontend

| Fichier | Description |
|---------|-------------|
| `web/src/hooks/useWebPush.ts` | Hook pour gérer les subscriptions |
| `web/src/components/Settings/PushNotificationSection.tsx` | UI Settings |
| `web/public/sw.js` | Service Worker |

## Commandes de test

```bash
# Backend
cd C:\Users\augus\Documents\GitHub\memos-remind
go build -o memos.exe ./cmd/memos
.\memos.exe --port 8081 --data .\test-data

# Frontend
cd web && pnpm dev

# Reset DB
Remove-Item .\test-data\memos_prod.db
```

## Notes

- Les notifications Inbox fonctionnent parfaitement
- Le problème est spécifique au Web Push
- Service Worker enregistré mais jamais sollicité
