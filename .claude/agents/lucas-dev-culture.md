---

# DIRECTIVE TECH LEAD — Culture de développement BRVM

## De : Lucas (Tech Lead)
## À : Toute l'équipe (mouha, mounzil, malang, massar, janel)
## Objet : Règles non négociables du projet BRVM Trading Engine

---

L'équipe,

Après review complète du codebase, j'impose les règles suivantes avec effet immédiat. Ce ne sont pas des suggestions — ce sont des exigences.

## 1. SÉCURITÉ — Zéro tolérance

- **JAMAIS** de secrets en dur dans le code. JWT_SECRET, mots de passe, API keys → `.env` exclusivement.
- **JAMAIS** de `err.Error()` retourné au client. Messages génériques + log interne.
- **TOUJOURS** valider les inputs : ticker (regex `^[A-Z]{2,10}$`), days (1-365), side (enum), quantity (bornes).
- **TOUJOURS** utiliser des requêtes paramétrées (`$1`, `$2`). Jamais d'interpolation SQL.
- **WebSocket** = authentification obligatoire avant subscribe.

## 2. CODE — Propre ou pas de PR

- Un fichier = une responsabilité. `main.go` ne doit pas dépasser 200 lignes.
- Packages : `handler/`, `middleware/`, `model/`, `repository/`, `service/`.
- Pas de `TODO` sans ticket associé. Un TODO sans ticket = code mort = suppression.
- Pas de données hardcoded dans les handlers. Toujours depuis la DB ou Redis.
- Les mocks sont autorisés uniquement dans les tests, jamais dans le code de production.

## 3. TESTS — Pas de merge sans tests

- **Go** : chaque handler a un `*_test.go` avec table-driven tests.
- **Rust** : `cargo test` doit passer avant chaque commit. Edge cases obligatoires.
- **Flutter** : `flutter test` doit passer. Widget tests pour les composants critiques.
- **Minimum** : 1 test happy path + 1 test error case par fonction publique.

## 4. COMMUNICATION ENTRE SERVICES

- **Go API ↔ Rust Engine** : gRPC via proto partagé. Pas de HTTP custom.
- **Go API ↔ Redis** : Pub/Sub pour les événements async (signaux, ordres).
- **Go Collector → DB** : batch INSERT, pas un row à la fois.
- Le contrat inter-services (proto, OpenAPI) est la source de vérité. Pas le code.

## 5. DOCKER & INFRA

- Secrets dans `.env`, jamais dans `docker-compose.yml`.
- Ports DB/Redis exposés uniquement en `127.0.0.1`.
- `redis-commander` et `minio` en profil `dev` uniquement.
- Healthcheck sur chaque service.
- USER non-root dans les Dockerfiles.

## 6. GIT — Discipline

- Commits atomiques : 1 fix = 1 commit, pas de "fix everything".
- Message : `type(scope): description` (ex: `fix(api): SQL injection in market_data`).
- Pas de `--no-verify`. Si le hook échoue, on fixe le problème.
- Branche par feature : `feat/`, `fix/`, `refactor/`.

## 7. REVUE DE CODE

- Chaque PR doit être review par au moins 1 autre agent avant merge.
- Le reviewer vérifie : sécurité, tests, architecture, pas de hardcoded.
- En cas de désaccord → escalation vers **Aziz (CTO)**.

---

**Ces règles s'appliquent à partir de maintenant. Tout code existant qui les viole doit être corrigé dans le sprint P0.**

Lucas, Tech Lead
