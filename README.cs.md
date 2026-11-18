# Oprava Otevřít, verze 0.1.0-0003

Tlačítko v Centru balíčků již není propojeno přes INFO dsmappname se spouštěčem v /3rdparty.
INFO nyní používá adminprotocol=https, adminport=18443, adminurl="".
Položka nabídky DSM zůstává zachována přes ui/config a HTML přesměrování.
Odkaz URL v Centru balíčků nyní vede na kořen aplikace; adresa instalačního
kódu zůstává v popisu balíčku a je na cestě /_synology/setup na HTTPS portu 18443.

Nahraď zdroje, sestav přes GitHub Actions a aktualizuj ruční instalací SPK.
Neodinstalovávej balíček, aby se zbytečně neriskovala ztráta dat.
Po aktualizaci ověř verzi 0.1.0-0003 a odhlas/přihlas DSM pro obnovení nabídky.
Minimum hesla 8 znaků i automatické ikony z icon/icon.png jsou zachované.

Ověřena syntaxe upraveného Python skriptu a přítomnost nastavení v kódu.
Chování tlačítka v DSM ani kompilace nebyly v tomto prostředí otestovány.

## Změny verze 0.1.0-0002

- Minimum hesla instalačního portálu: 8 znaků ve všech třech kontrolách.
- Obrázek vlož do `icon/icon.png`. Vygenerovaný obrázek je již přiložený.
- Sestavení automaticky vytvoří ikony Centra balíčků i nabídky DSM.
- Otevřít přesměruje na HTTPS, stejnou adresu NAS a port 18443.
- Původní heslo již nainstalovaného portálu se aktualizací nemění.
- Pro aktualizaci existující instalace je zvýšena verze na 0.1.0-0002.

Postup: nahraj obsah do GitHubu včetně `.github`, spusť Actions a stáhni artefakt.
Ikonu stačí nahradit před sestavením; na NAS nic ručně nekopíruješ.
Minimum 8 znaků není doporučení používat slabé heslo. Používej jen v důvěryhodné LAN.

# Prusa Connect Local pro Synology DSM 7

Experimentální zdrojový prototyp pro DS220+. **Nejde o otestovaný instalační balíček.**
Archiv obsahuje zdrojové kódy a automatické sestavení SPK přes GitHub Actions.
Podrobnosti a omezení jsou v README.md.

## Instalace bez SSH a Plánovače úloh

1. Nahraj obsah archivu do vlastního GitHub repozitáře včetně složky `.github`.
2. V Actions spusť `Build experimental DSM7 package` tlačítkem `Run workflow`.
3. Po úspěšném sestavení stáhni artefakt a rozbal soubor `.spk`.
4. V DSM otevři Centrum balíčků → Ruční instalace. Vyber SPK.
5. Zvol jedinečné náhodné heslo instalačního portálu, alespoň 8 znaků.
6. Otevři balíček. Spouštěč automaticky převezme adresu a port aktuálního DSM.
7. Odkaz na instalační kód otevře HTTPS portál. Uživatel je `admin`, heslo je z kroku 5.
8. Zkopíruj jednorázový kód a dokonči nastavení v Prusa Connect.

HTTPS aplikace používá port 18443 a vlastní certifikát, kterému prohlížeč automaticky
nedůvěřuje. HTTP port 18080 slouží pro připojení tiskáren, ale vystavuje i API aplikace.
Používej jen v důvěryhodné LAN. Balíček nemění firewall ani router. Je-li firewall
zapnutý, povol potřebné porty v grafickém rozhraní DSM jen pro svou LAN.

URL kódu je v popisu balíčku: `https://<adresa-NAS>:18443/_synology/setup`.
Spouštěč vytvoří konkrétní odkaz automaticky. Kód není veřejný ani uložený do logu.
České a anglické je vlastní rozhraní; překlady Prusa Connect závisí na upstreamu.

Na skutečném NAS nebyla ověřena instalace, oprávnění, restart ani komunikace tiskáren.
Sestavení v GitHub Actions v této konverzaci nebylo spuštěno. Nejdřív vývojový test,
ne nasazení do produkce. Před aktualizací nebo odinstalací zálohuj data.
