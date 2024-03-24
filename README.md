# Projektarbeit: [Vault by HashiCorp](https://www.vaultproject.io/)
Vault ist ein System zur identitäsbasierten Verwaltung von Secrets und Verschlüsselung als zentrale Kontrolleinheit in Unternehmen und anderen großen Organisationen.
Secrets können dabei alle möglichen Formen von Daten sein, zu denen der Zugang streng kontrolliert werden soll, wie etwa Passwörter, Zertifikate, Schlüsseldateien für Datenverschlüsselung und vieles mehr.
Der Zugriff auf Vault ist über eine grafische Benutzeroberfläche, ein Kommandozeileninterface sowie eine HTTP-API möglich.
Für viele Programmiersprachen existieren Bibliotheken, die den Zugriff auf die HTTP-API kapseln und Vault somit nativ in die jeweilige Sprache einbinden.

Sowohl für Menschen als auch für Maschinen existiert eine Vielzahl an Möglichkeiten, um sich bei Vault zu authentifizieren.
Hier können über OIDC, LDAP, GitHub und Co. auch externe Identity Provider verwendet werden.
Somit ist es für Unternehmen einfach möglich, ihre vorhandenes Identitäts- und Access-Managementsystem an Vault anzubinden, wodurch die Mitarbeiter automatisch Zugriff erhalten.

Für jedes von Vault verwaltete Objekt ist eine feingranulare Zugriffskontrolle möglich, um zu steuern, wer welche Operationen auf einem Objekt ausführen darf.
Nach dem Prinzip der minimalen Berechtigungen ist es in Vault sehr einfach möglich, die Zugriffsrechte jedes Benutzers (oder einer Benutzergruppe) auf das nötige Minimalmaß zu beschränken.
Vault bietet nicht nur die Möglichkeit, den Zugriff auf Secrets identitätsbasiert zu beschränken, sondern protokolliert überdies detailliert jeden einzelnen Zugriff auf Secrets in einem Auditlog, um jederzeit transparent nachvollziehen zu können, wer wann auf was zugegriffen hat.

Use Cases für Vault inkludieren unter anderem:
- die allgemeine Verwaltung von Secrets aller Art
- Encryption as a Service (Bereitstellung von High Level APIs zur Ver- und Entschlüsselung von Daten)
- die Verwaltung von Secrets in Kubernetes
- Credential Rotation für Datenbankzugangsdaten
- Dynamic Secrets
- Identitätsbasierte Zugriffsverwaltung
- uvm.

<!-- ## Use Cases
- Verwaltung von Secrets
- Dynamische Secrets
	- on demand generiert und eindeutig für einen Client
- Kubernetes Secrets
- Database Credential Rotation
- Automated PKI Infrastructure
- Identity-based Access
- Datenverschlüsselung und -Tokenisierung
- Schlüsselverwaltung -->

Ein besonderes Feature von Vault, welches das Secret Management System von anderer Software absetzt, sind sogenannte *Dynamic Secrets*.
Dynamic Secrets sind Zugangsdaten, die on demand individuell für einen Client generiert werden und üblicherweise eingeschränkte Berechtigungen sowie eine kurze TTL (Time to Live) haben.

Ziel dieser Projektarbeit ist die Erkundung von Vault, wobei ein spezieller Fokus auf Dynamic Secrets im Rahmen von Datenbankzugriffen gelegt wird.
Dafür wird ein kleines Beispielprojekt erstellt und beschrieben, in dem einige Use Cases und Features von Vault demonstriert werden können.

## Laborumgebung
Häufig ist insbesondere in Cloud-Umgebungen der Einsatz von Vault von Vorteil.
Daher wird dieses Projekt mit dem de facto Standard für Containerorchestrierung - Kubernetes - realisiert.
Dadurch wird ein realitätsnahes Test-Umfeld geschaffen und eine gute Reproduzierbarkeit des Projekts ermöglicht.
Getestet wurde das Projekt auf einem [SAP Gardener Cluster](https://gardener.cloud/).
Da jedoch keine komplexen, Plattform-Provider-spezifischen Funktionen von Kubernetes benutzt werden, ist davon auszugehen, dass das Projekt auch auf Plattformen anderer Hyperscaler oder sogar lokal (z.B. mit [Minikube](https://minikube.sigs.k8s.io/docs/)) reproduziert werden kann.
Die benötigten externen Abhängigkeiten, darunter auch Vault selbst, werden mithilfe von [Helm]() provisioniert.
Zur reproduzierbaren Konfiguration der Laborumgebung werden Bash-Scripts verwendet.

Kernstück des Beispielprojekts ist die **Dynamic Secrets** Funktion von Vault.
Zur Demonstration wird hier eine [PostgreSQL-Datenbank](https://www.postgresql.org/) verwendet, aus der ein Service im Kubernetes Cluster Daten liest und diese als Website bereitstellt.
Die erforderliche Konfiguration der Datenbank wird mittels eines [Go](https://go.dev/)-Programms realisiert.
Auch der Service zum Lesen der Daten und zur Darstellung als Website sind in Go geschrieben.

## Installation
> Voraussetzung für die folgenden Schritte ist die lokale Installation von [kubectl](https://kubernetes.io/de/docs/tasks/tools/install-kubectl/) und [Helm](https://helm.sh/), sowie das Herstellen einer Verbindung zu einem Kubernetescluster.

Zunächst werden Vault und die PostgreSQL-Datenbank im Kubernetes-Cluster installiert.
Die notwendigen Schritte werden von [create-landscape.sh](scripts/create-landscape.sh) ausgeführt.
Dabei werden als erstes die Repositories von Vault und PostgreSQL als Quellen in Helm hinzugefügt.
In Kubernetes wird für jede Anwendung ein eigener *Namespace* angelegt.
Dadurch entsteht eine saubere Trennung und die Resourcen können leicht der entsprechenden Anwendung zugeordnet werden.
Anschließend werden die beiden Andwendungen mit Helm installiert.
Die individuelle Anpassung der Konfigurationen erfolgt dabei über die jeweilige Konfigurationsdatei ([*vault-postgre.yaml*](values-postgre.yaml) / [*values-vault.yaml*](values-vault.yaml)).
Für dieses Beispielprojekt kann größtenteils die Standardkonfiguration verwendet werden.
Jedoch wird für PostgreSQL das Standardpasswort zur Administration gesetzt und für Vault eine [Ingress-Adresse](https://kubernetes.io/docs/concepts/services-networking/ingress/) gesetzt, über die die Benutzeroberfläche von Vault anschließend im Internet erreichbar ist.

## Bootstrap der Datenbank
Für das Beispielprojekt und die Integration in Vault ist eine initiale Konfiguration der PostgreSQL-Datenbank erforderlich.
Durchgeführt wird diese mittels eines [Go-Programms](db-setup/main.go).
Da die im Kubernetes-Cluster gehostete Datenbank in ihrer aktuellen Konfiguration nicht ins Internet geroutet wird, bieten sich für die Kommunikation des Bootstrap-Programms mit der Datenbank zwei Möglichkeiten.
Die einfacher umzusetzende Variante, die jedoch mehr manuelle Intervention erfordert, ist es, über kubectl ein Port-Forwarding der Datenbank auf das lokale System herzustellen und das Go-Programm so auszuführen, als würde die Datenbank auf demselben Rechner ausgeführt.
Die Alternative ist die Verwendung eines Containers zur Initialisierung.
Dabei wird das Go-Programm containerisiert und direkt im Kubernetes-Cluster ausgeführt, wodurch es direkten Netzwerkzugriff auf die Datenbank erhält.
Nachteil dieses Ansatzes ist der deutlich größere Overhead, da nicht nur das Programm zum Bootstrap entwickelt, sondern zusätzlich noch ein [Containerfile](db-setup/Containerfile) für die Paketierung der Applikation geschrieben und die Anwendung auf eine Container-Registry gepusht werden muss, von der das Kubernetes-Cluster den neu erstellten Container beziehen kann.
Da dieser Prozess jedoch ohnehin für den Webserver durchlaufen werden muss, ist die zusätzliche Übertragung auf das Programm zum Bootstrap der Datenbank kein großer Aufwand.

Im Bootstrap werden initiale Daten (eine Tabelle mit mehreren Einträgen) angelegt, die der Webservice später anzeigen soll.
Weiterhin wird eine Nutzerrolle angelegt, welche den Zugangsdaten der zugreifenden Applikation zugewiesen werden kann.
Mit dieser Rolle werden die Leserechte auf ein bestimmtes Schema der Datenbank eingeschränkt und verwehrt sämtliche Schreibrechte.
Die gewählten Einschränkungen sind zu Demonstrationszwecken grob gehalten und könnten in der Praxis noch deutlich feiner und restriktiver gewählt werden.
Mit der angelegten Rolle kann Vault die Zugriffsrechte von Dynamic Secrets steuern.
Voraussetzung ist es jedoch, dass die entsprechenden Rollen bereits in der Datenbank existieren und konfiguriert sind.

<!-- - Exposen von DB via Ingress nicht einfach möglich, weil kein HTTP -->

## Konfiguration von Vault
Mittels [setup-vault.sh](scripts/setup-vault.sh) wird die initiale Konfiguration für dieses Beispielprojekt auf Vault eingerichtet.
Dazu werden zunächst verschiedene Policies angelegt.
Über Policies steuert Vault die Zugriffsrechte der jeweiligen Benutzer und Anwendungen.
Die Struktur von Vault ist grundsätzlich so aufgebaut, dass Funktionalitäten jeglicher Art einen fest zugewiesenen Pfad haben.
Die in [HCL (HashiCorp configuration language)](https://github.com/hashicorp/hcl) geschriebenen Policies spezifizieren jeweils für eine Reihe an Pfaden, welche Operationen (ähnlich [CRUD-Operationen](https://www.crowdstrike.de/cybersecurity-101/observability/crud/)) der Rechteinhaber auf ihnen ausführen darf.
Da alle Ressourcen innerhalb von Vault solch einen Pfad haben, steuern Policies so sowohl den Zugriff auf Secret Engines als auch auf die Konfiguration von Vault selbst.
Für das Beispielprojekt werden drei Policies angelegt: eine *Admin-Policy* mit allen Rechten, eine *Global-Reader-Policy* mit Leserechten für alle Pfade und eine Policy für den *Webservice*, welche einzig Leserechte auf die Dynamic Secrets der Datenbank Zugriff erhält.

Anschließend werden zur Demonstration verschiedene Authentifizierungsmethoden aktiviert.
Einerseits wird eine herkömmliche Authentifizierung mit Benutzernamen und Passwort (*Userpass*) aktiviert, um menschliche Benutzer nachzustellen.
Für diese Authentifizierungsmethode werden drei verschiedene Benutzer erzeugt, welche jeweils unterschiedliche Zugriffsrechte besitzen.
Einen Administrator, einen Global-Reader und einen Nutzer, dem keine expliziten Rechte zugewiesen werden.  
Andererseits wird die *Approle*-Authentifizierungsmethode aktiviert, um eine Authentifizierung des Webservice zu ermöglichen.
Vault unterstützt darüber hinaus noch viele weitere Authentifizierungsmethoden, die jedoch deutlich aufwändiger zu konfigurieren sind, beziehungsweise erst in Enterprise-Netzwerken mit bereits vorhandenen Authentifizierungsstrukturen ihre Wirksamkeit entfalten können.

Für die Approle-Authentifizierungsmethode gibt es viele Optionen, um den Zugriff weiter einzuschränken, die jedoch zur Vereinfachung der Demonstration hier nicht verwendet werden.
Im Kontrast zur Userpass-Authentifizierungsmethode können nicht direkt Zugangsdaten angelegt werden, sondern es muss erst eine Rolle angelegt werden, von der anschließend Zugangsdaten abgeleitet werden können.
Das Script legt eine Rolle *webviewer* an und erzeugt im Anschluss direkt Zugangsdaten, die der Webservice zur Authentifizierung verwenden soll.
Ähnlich zu Benutzername und Passwort besteht eine Approle aus einer *Role-ID* und einer *Secret-ID*.
Diese werden als [Kubernetes Secret](https://kubernetes.io/docs/tasks/inject-data-application/distribute-credentials-secure/) angelegt, um sie dem Pod des Webviewers bereitstellen zu können.

Anschließend wird die Datenbank-Secret-Engine aktiviert und konfiguriert.
Dazu wird erst eine Datenbankverbindung angelegt, welche die Master-Zugangsdaten der Datenbank verwendet.
Sobald die Verbindung eingerichtet ist, kann Vault die volle Kontrolle über die Master-Zugangsdaten übernehmen und diese automatisiert rotieren, um Angriffspunkte zu minimieren.
Sobald die erste Rotation durchgeführt wurde, sind die Master-Zugangsdaten nicht mehr manuell erreichbar.
Ähnlich wie bei Approles werden in der Datenbank-Secret-Engine verschiedene Rollen angelegt, aus welchen dann die eigentlichen Zugangsdaten für die Datenbank erstellt werden können.
Einer Rolle wird primär eine Datenbank zugewiesen, sowie ein SQL *Creation Statement*, welches auf der Datenbank zur Erzeugung eines Accounts mit den entsprechenden Rechten ausgeführt wird.
Weiterhin können in der Rolle Parameter wie die TTL der Dynamic Secrets festgelegt werden.

## Webviewer
Der Webviewer ist ein simpler [Go-Service](webviewer/main.go), welcher Daten aus einer Tabelle der PostgreSQL-Datenbank liest und diese als HTML-Website darstellt.
Dabei benutzt er die Approle-Authentifizierungmethode von Vault, um ein Dynamic Secret anzufragen, mit dem dann wiederum auf die Datenbank zugegriffen werden kann.

Nach dem Start baut der Webviewer zunächst über `renewConnection()` eine Verbindung zur Datenbank auf.
Dabei findet in der `getDtabaseCredentials()`-Funktion der eigentliche Authentifizierungsprozess mit Vault statt.
Zur Kommunikation mit Vault wird das von HashiCorp bereitgestellte [API-Paket](https://pkg.go.dev/github.com/hashicorp/vault/api@v1.12.2) für Go verwendet.
Alternativ könnte der Zugriff auch über die HTTP-API erfolgen.
Letzteres ist der universelle Weg, über den sämtliche Programmiersprachen (Unterstützung für HTTP-Kommunikation vorausgesetzt) mit Vault kommunizieren können.
Die Verwendung der HTTP-API bringt jedoch eigenen Herausforderungen mit sich - nicht zuletzt ist eine Fehlerprüfung durch den Compiler nur bedingt möglich.  
Im Webviewer wird eine Standardkonfiguration von Vault geladen und um die korrekte Adresse ergänzt, unter der Vault zu erreichen ist.
Anschließend wird ein Vault-Client mit dieser Konfiguration initiiert.
Dieser Client kann im aktuellen Zustand jedoch noch keine Daten aus Vault abrufen, sondern muss erst die Authentifizierung durchführen.
Die für die Authentifizierung benötigten Approple-Zugangsdaten werden dem Go-Programm über Umgebungsvariablen bereitgestellt.
So ist es möglich, den Container zu bauen ohne sich auf bestimmte Zugangsdaten festlegen zu müssen oder Gefahr zu laufen, diese mit dem Container-Image offenzulegen.
Anschließend führt der Client eine Authentifizierung mit den Approle-Zugangsdaten durch.
Der nun authentifizierte Client kann jetzt ein Dynamic Secret für den Zugriff auf die Datenbank anfordern und dieses zurückgeben.

Mit diesem Dynamic Secret kann nun von `connectDB()` die eigentliche Verbindung zur Datenbank aufgebaut werden.
Dabei wird die [native Datenbank Schnittstelle](https://pkg.go.dev/database/sql) von Go mit einem Treiber für [PostgreSQL](https://pkg.go.dev/github.com/lib/pq@v1.10.9) verwendet.
Als Nächstes lädt der Webviewer die [Templates](https://pkg.go.dev/html/template) für die eigentliche Website und eine Seite für Fehlermeldungen.
Mittels Templates können sehr einfach Datenstrukturen aus Go in sicheres HTML gerendert werden.
Durch ein initiales Parsen der Templates müssen diese nicht bei jeder Anfrage an den Webserver neu verarbeitet werden.
Der Webviewer registriert `genPage()` als zentralen Handler für Anfragen und geht in den Serving-Mode.

Wir nun der Webviewer von einem Client angefragt, wird `genPage()` aufgerufen.
Die Funktion startet eine Query, um die Datenbanktabelle auszulesen, rendert die Ergebnisse über das HTML-Template und schreibt die resultierende Website in den ResponseWriter.
Im Fall, dass die Kommunikation mit der Datenbank fehlschlägt, wird die Fehlermeldung mit dem Error-Template gerendert.
Weiterhin enthält diese Fehlerseite einen Link, um die Datenbankverbindung neu aufzubauen.
Dieser kann benutzt werden, falls das Dynamic Secret abgelaufen ist oder widerrufen wurde, um die Funktionalität des Webviewers wiederherzustellen.

Der Webviewer wird mittels eines [Containerfiles](webviewer/Containerfile) zu einem Docker-Image gebaut und auf eine Containerregistry gepusht, von wo Kubernetes es herunterladen kann.
Um die Anwendung in Kubernetes zu provisionieren, wird eine [YAML-Datei](deployments/webviewer-deployment.yaml) verwendet, um das Deployment reproduzierbar zu beschreiben.
Beschrieben werden drei Kubernetes-Objekte:
- ein Deployment
- ein Service
- ein Ingress

Mittels des Deployments wird das eigentliche Containerimage als Pod in Kubernetes ausgeführt.
Außerdem werden dabei die Umgebungsvariablen für die Zugangsdaten der Approle-Authentifizierung aus dem Kubernetes-Secret geladen, welches bereits beim Setup-Prozess von Vault angelegt wurde.
Diese Umgebungsvariablen stehen dann dem Webviewer zur Verfügung.
Durch das Laden der Variablen aus dem Kubernetes-Secret wird vermieden, die Zugangsdaten unverschlüsselt übergeben zu müssen.
Um auf den Webviewer zugreifen zu müssen, wird ein Service angelegt, welcher den Zugriff auf den Webserver im Kubernetes-Cluster freigibt.
Mittels dieses Servers ist es dann möglich, ein Ingress-Objekt zu erzeugen, welches den Webviewer im Internet bereitstellt.
Das Deployment muss abschließend noch mit `kubectl apply -f deployments/webviewer-deployment.yaml` auf Kubernetes erzeugt werden.

---

## Showcases in Laborumgebung
- Login mit verschiedenen Nutzern; Demonstrieren unterschiedlicher Berechtigungen
- Pfade zu verschiedenen Secret Engines erkunden
  - Secret in KV Secret Engine anlegen; kann mit Nutzer A modifiziert werden, mit Nutzer B nur angezeigt werden, Nutzer C sieht es nicht
- Dynamic Secret für Datenbank im Webinterface abrufen und manuell über Kommandozeile in PostgreSQL einloggen und Datenbankbenutzer anzeigen
  - `kubectl exec -n postgre -it postgre-postgresql-0 -- /bin/bash`
  - `env PGPASSWORD=<password> psql -d postgres -p 5432 -U <user>`
  - `\du`
- Dashboard *Quick Actions* kann das selbe deutlich schneller
- Webviewer deployen; keine Zugangsdaten im Code, trotzdem kann Inhalt der Datenbank angezeigt werden
  - Zugriff der Anwendung im Auditlog anzeigen
  - in Datenbank den neu angelegten Nutzer anzeigen
  - Zugangsdaten widerrufen (s.u.)
- Zugangsdaten mit Benutzer in Webinterface abrufen und im Auditlog Zugriff nachvollziehen (`kubectl logs -f -n vault vault-0`)
- Verschlüsseln von Daten mit Transit Engine


## Zugangsdaten widerrufen
In der Ausgangslage zeigt der Webviewer die Tabelle der Datenbank ordnungsgemäß an.
Die Rollentabelle von Postgres führt den Nutzer des Webviewers `v-approle-readonly-IRPzIp8tHztbnbMLVAl2-1711209306`.
Als Dynamic Secret besitzt dieser Nutzer lediglich eine beschränkte Gültigkeit.
```
postgres=# \du
                                                  List of roles
                     Role name                      |                         Attributes                         
----------------------------------------------------+------------------------------------------------------------
 postgres                                           | Superuser, Create role, Create DB, Replication, Bypass RLS
 ro                                                 | No inheritance, Cannot login
 v-approle-readonly-IRPzIp8tHztbnbMLVAl2-1711209306 | Password valid until 2024-03-23 16:55:11+00
```
Werden nun in Vault alle Dynamic Secrets (mit [`scripts/revoke-all-db-creds.fish`](scripts/revoke-all-db-creds.fish)) widerrufen, löscht Vault diese umgehend aus der Datenbank: 
```
postgres=# \du
                             List of roles
 Role name |                         Attributes                         
-----------+------------------------------------------------------------
 postgres  | Superuser, Create role, Create DB, Replication, Bypass RLS
 ro        | No inheritance, Cannot login
```
Dies könnte beispielsweise als Schutzmaßnahme durchgeführt werden, falls bemerkt wird, dass Zugangsdaten für die Datenbank kompromittiert wurden.  
In der Folge sind die Master-Zugangsdaten, die nur Vault zur Verfügung stehen, die einzigen gültigen Zugangsdaten.
Alle geleakten Zugangsdaten sind somit wertlos und stellen keine Gefahr mehr da.

Berechtigte Anwendungen können sich anschließend ganz einfach von Vault neue Zugangsdaten generieren lassen und mit diesen ihre Operationen fortführen.
```
postgres=# \du
                                                  List of roles
                     Role name                      |                         Attributes                         
----------------------------------------------------+------------------------------------------------------------
 postgres                                           | Superuser, Create role, Create DB, Replication, Bypass RLS
 ro                                                 | No inheritance, Cannot login
 v-approle-readonly-S24a9X3cxSjRRZiGcfFf-1711209344 | Password valid until 2024-03-23 16:55:49+00
```

## Probleme und Vereinfachungen in der Laborumgebung
- überall ist SSL ausgeschaltet
  - entscheidende Sicherheitslücke: HTTPS wird ausgehebelt, für Produktionssystem natürlich untragbar
  - macht Entwicklung bedeutend simpler, aber natürlich großes Sicherheitsrisiko
- Vault läuft nur im Dev-Modus
  - viele Hardening-Mechanismen komplett ausgehebelt
  - Vault automatisch unsealed, benötigt nicht mehrere Keys oder Trusted Authority
  - Daten nur in Memory
  - ...
- *Global Reader* erhält automatisch Zugriff zu sämtlichen Datenbankrollen
  - -> Policies müssen sehr umsichtig geschrieben werden, "read" bedeutet nicht automatisch nur lesen
- Approle-Authentifizierung
  - ist so zwar bereits gut, weil App selbst die Credentials nicht kennt und diese erst beim Deployment in Kubernetes verknüpft werden
  - ABER: könnte [Response Wrapping](https://developer.hashicorp.com/vault/docs/concepts/response-wrapping) bei der Authentifizierung verwenden (nicht das eigentliche Secret wird übertragen, sondern ein "single-use" Token, mit dem das eigentliche Token einmalig unwrapped werden kann)
  - Lebenszeit von Approle-Credentials begrenzen
  - Verwendung von [Vault Agent](https://developer.hashicorp.com/vault/docs/agent-and-proxy/agent) zur Verteilung von Credentials direkt in Kubernetes
- Bootstrap-Container der Datenbank enthält Master-Zugangsdaten der Datenbank und ist zudem auf öffentlicher Container Registry erhältlich
  - spiegelt selbstverständlich keinen realen Produktionseinsatz wider
  - in der Praxis würde dieser Vorgang höchst wahrscheinlich auch nicht automatisiert stattfinden, sondern die Umgebung würde manuell angelegt werden
  - hier nur dieser Ansatz gewählt, um die Umgebung einfach und automatisch reproduzieren zu können
- Webviewer führt keine automatische Erneuerung von Credentials durch
  - Code kann automatisch mit *LifetimeWatcher* für Erneuerung der Credentials sorgen
  - deutlich größerer Overhead, daher hier nicht realisiert
- "Bug" in Global-Reader Rolle
  - *Read*-Operation auf alle Pfade erlaubt
  - Dynamic Secrets für Datenbanken werden mit *Read*-Call generiert
  - daher kann ein Global-Reader in dieser Implementierung auch selbst neue Credentials erzeugen und geht über einen reinen Lesezugriff streng genommen hinaus

## Gewonnene Erkenntnisse zu Vault
### Positive Aspekte
+ sehr feingranulare Einstellungsmöglichkeiten
+ nach vollständiger Implementierung: ermöglicht es sehr einfach, den Lebenszylkus von Zugangsdaten vollautomatisiert zu steuern
+ durch Dynamic Secrets können Key-Rotations in vielen Bereichen vollständig automatisiert durchgeführt werden
+ Secret Management und Verschlüsselung sind häufig verantwortlich für Schwachstellen in Systemen. Mit Vault müssen Entwickler weniger dieser Aufgaben selbst übernehmen, wodurch die Fehlerwahrscheinlichkeit sinkt und die Einfachheit des Codes idealerweise verbessert wird
+ deckt ["so ziemlich alle Anwendungsfälle und Notwendigkeiten einer grösseren Organisation in Sachen Secrets und Secret Management"](https://b-nova.com/home/content/heres-how-easy-effective-and-secure-secrets-management-is-using-hashicorp-vault/) ab
+ Open Source Version (bereits mit sehr großer Bandbreite an Features) ohne Gebühren auch in großen Organisationen nutzbar
  + optional Enterprise-Version mit weiteren Funktionen erhältlich
  + kann auch als SaaS-Dienst von HashiCorp gebucht werden
+ "Eierlegende Wollmilchsau": aufgrund der vielen Authentifizierungsmethoden, Secret Engines und Plugins, kann Vault extrem vielseitig eingesetzt werden

### Negative Aspekte
- Token-Lifecycle unter Umständen sehr aufwändig zu implementieren
  - allerdings: Wie sähe es erst ohne Vault aus?
- Einstellungen müssen gut durchdacht werden
  - siehe Global-Reader "*Bug*"
  - es ist unter Umständen leicht, unbedachte Konfigurationen anzulegen
- Einstiegshürde für spezifischere und komplexere Konfigurationen
  - rein textbasierte Konfiguration und Policies
  - nur bedingte Anleitungen abgesehen von Tutorials
  - CLI sehr unintuitiv
  - außer über UI muss man normalerweise sehr genau wissen, welchen Pfad man anfragen möchte

## Alternativen
Alternativen zu Vault sind schwer zu vergleichen, da die primären Informationsquellen, die zur Verfügung stehen, auf den Werbeinformationen der Konkurrenten beruhen.
Nennenswerte konkurrierende Produkte, die häufig Erwähnung finden, sind unter anderem:

### Lösungen für Secret Management von Hyperscalern
  - AWS Secrets Manager
  - Azure Key Vault
  - Google Cloud Secret Manager

Solche Lösungen sind in der Regel für den Einsatz in ihrem spezifischen Umfeld optimiert und können nicht so vielseitig integriert werden wie Vault.

### Infisical
- aufstrebende Open Source Lösung mit ähnlichen Funktionen wie Vault
- Vault sei mit mehr manuellem Aufwand verbunden, aber dafür strukturell individueller anpassbar

Ein Herausstellungsmerkmal von Vault sind vor allem die vielseitigen Einsatzmöglichkeiten.
Außerdem geht der Ansatz von HashiCorp über ein reines Speichern von Secrets hinaus zu einem Secret Management, welches den kompletten Lebenszyklus von Secrets steuert.

---
## Hinweis
Aufgrund von sensiblen Daten wurden die Ingress-Adressen und der Dateipfad der Kubeconfig-Datei mit Platzhaltern *`REPLACE_XY`* ersetzt.
