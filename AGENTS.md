# mizan-go: ajan kuralları

Bu depoda çalışan her ajan (ve insan) için kurallar. Mizan'ın görsel dilini MyGo'nun yerel arayüzüne (`ui` paketi) taşır; davranış (odak, klavye, erişilebilirlik) MyGo'nun görünümsüz `…Base` bileşenlerinden gelir, Mizan görünümü verir. Bir kural bozulacaksa nedeni koda yorum olarak yazılır ve bu dosya güncellenir.

## 1. Yapı ve yazım

- Her bileşen ailesi kök dizinde bir dosya: `controls.go`, `forms.go`, `overlays.go`, `dnd.go`, `feedback.go`, `disclosure.go`, `toggle.go`… Renk, gölge ve boyut değerleri `tokens.go`'da, OKLCH ile.
- Yüzey dili iki tür: kabarık denetim (`t.raised`, ışık yukarıdan) ve gömülü kuyu (`t.well`). Bileşen kendi gradyanını yazmaz, bu ikisini kullanır.
- Kod yorumları İngilizce, kararı anlatır, yoğunluğu çevredeki kodla aynı. Dışa açık her şeyin belge yorumu bir kullanım örneği içerir.
- Paket erken beta: API serbestçe değişir, ama değişince README ve gallery aynı turda güncellenir.
- Her yeni bileşen gallery'de (`gallery/`) görünür ve `gallery/*_test.go`'da etkileşim testi alır.
- Kapı: `gofmt -l .` boş, `go vet ./...`, `go test ./...`, ve CGO'suz `linux`, `darwin`, `windows`, `js/wasm` derlemesi (CI ile aynı).

## 2. Hareket ve mikro etkileşim

Kullanıcı için küçük geri bildirimler ve animasyonlar çok önemli. Bir etkileşim "çalışıyor" diye bitmez; hissi de doğru olmalı.

### İlkeler

- **Her durum değişikliği hareketle gelir.** Açılan, kapanan, yer değiştiren, eklenen ya da silinen hiçbir şey bir karede belirip kaybolmaz. Kapanış da canlandırılır (`presence` kalıbı, aşağıda).
- **Gelirken yay, giderken eğri.** Giriş yayla (`ui.SpringOptions`), biraz sekmeyle; çıkış daha kısa bir eğriyle (`EaseOut`, 150–220 ms) ve sekmesiz. Çıkış girişten hızlıdır.
- **Sekme ölçülü.** Konum değiştiren şeylerde `Bounce` 0.04–0.16; değer gösteren şeylerde (Progress dolumu, yer açan kartlar) 0: değerinin ötesine geçen çubuk, olduğundan fazla okunur.
- **Ters yöne kaçma yok.** Bir öğe hedefine giderken önce geri gidip sonra ilerlemez. Konum ve genişlik birlikte değişiyorsa aynı yayla sürülür (Switch düğmesi: `x` ve `w` aynı yayda).
- **Ani sıçrama yok.** Bir değer iki farklı yoldan yazılıyorsa (yerleşim değişimi + transform) toplamı kare kare sürekli olmalı. Yeni yerleşime geçtiğin karede transform'u aynı miktarda geri al (`Jump`), sonra sıfıra yürüt (FLIP).
- **Fiziksel his.** Tutulan şey kalkar (biraz büyür, gölgesi `ShadowPop` olur), hıza göre hafifçe eğilir (`DragVelocity`, ±2–2.5°), bırakılınca yerine süzülür. Çevresindekiler ona yer açar; mavi çizgi gibi işaretlerle yetinilmez.
- **Basınca tepki.** Basılan düğme 1px iner ya da %96–98'e küçülür; tutulan düğme (stepper) önce bir adım atar, 400 ms sonra hızlanarak tekrarlar.
- **Ortamın canlı ama sakin olması.** Bir hedef "seni alırım" diyorsa düz renk yerine yavaşça yaşayan bir şey olabilir: kayan bulut gölgesi gibi `noise` ile dolaşan yumuşak lekeler (`cloud`). Çok hafif, çok yavaş (dönem birkaç saniye), dikkat dağıtmaz.
- **İşaretler çizilir.** Onay işareti belirmez, kalemle çizilir gibi kısa koldan uzun kola akar (`checkMark`, `Path.Trim` 0→1, 260 ms); kaldırılırken geldiği yoldan, daha hızlı ve dolguyla birlikte solarak geri çekilir. Açılan listelerde (Select, Combobox, Menu) seçili öğenin işareti liste açıldıktan ~90 ms sonra çizilir: seçim yapınca liste kapandığından değişimin kendisi pek görünmez.
- **Radio noktası açılır.** Seçilince dolgu önce gelir, nokta ~40 ms sonra merkezden sekmeden açılır (`dotMark`, 240 ms); bırakılınca 90 ms'de küçülüp söner. Yaylı "pop" kullanılmaz.
- **Onay ışıkları ince.** Bir işlem tamamlanınca kenarda koşan ışık (`rise`): 1px çekirdek, dar parıltı, ~1.5 s, `EaseInOut`, tek eğri (animasyonun eğrisini yolda ikinci kez eğme).
- **Blur yok.** Dialog ve Sheet'te bulanık hale kullanılmaz; yerine solup gelen `Scrim`. Cam (`glass`) yalnız popover, menü, tooltip ve toast'ta.
- **Az hareket isteği.** MyGo `AnimateTo` ve `Keyframes` `ReduceMotion`'da hedefe hemen gider; elle zaman sayan çizimler (`p.Now()`) de bunu gözetmeli.
- **Reddedilenler:** NumberInput'ta NumberFlow tarzı yuvarlanan rakamlar istenmedi (kaldırıldı). Benzer "sayaç dönmesi" önerileri getirme.

### Ayarlı yaylar (yeniden kullan)

| Ad | Nerede | Değer |
| --- | --- | --- |
| `springPanel` | Dialog girişi | 300 ms, bounce 0.12 |
| `springSheet` | Sheet girişi | 380 ms, bounce 0.06 |
| `springPop` | Popover, Menu | 220 ms, bounce 0.1 |
| `springPill` | Segmented hapı, Tabs alt çizgisi | 280 ms, bounce 0.14 |
| `springShift` | Yer açan kartlar | 320 ms, bounce 0.06 |
| `springLand` | Bırakılan kartın inişi | 380 ms, bounce 0.16 |
| `springLift` | Kartın kalkışı | 260 ms, bounce 0.2 |
| `stackSpring` | Toast yığını | 380 ms, bounce 0.04 |
| `springBar` | Progress dolumu | 500 ms, bounce 0 |

Yeni bir hareket önce bunlardan birini dener; yenisi gerekirse adlandırılıp buraya eklenir.

## 3. MyGo tuzakları (öğrenilmiş)

- **`Transition.Enter` yalnız ebeveyni önceki karede var olan öğede çalışır.** Dialog paneli backdrop ile aynı karede doğar, bu yüzden girişi hiç oynamaz; kapanınca backdrop'la birlikte anında gider. Çözüm `overlays.go` › `presence`: durum yerleşimde yer kaplamayan bir "host" öğede tutulur, ilerleme `Jump(0)` + `AnimateTo(1)` ile elle sürülür, katman çıkışı bitene dek çizilmeye devam eder ve kapanırken işaretçiyi alttaki sayfaya geçirir (`PassThrough`, `IgnoreInput`).
- **`AnimateTo` ilk çağrıda hedefte başlar.** Bir şeyi 0'dan canlandırmak için ilk karede `Jump(key, 0)` (bir `ui.Local` bayrağıyla tek kez).
- **Aynı öğeye iki `Translate` yazılırsa sonuncusu kazanır.** İki kaynaktan gelen hareket (liste kaydırması + iniş) ayrı öğelere konur: dıştaki yuva, içteki kart (`dnd.go` › `dragSource`).
- **`LayoutBounds` bir önceki karenin yerleşimidir.** Sıra değiştiren bir işlem, onu yeni sırayla çizen karenin *sonraki* karesinde ölçülür. Durumu değiştiren şey (ör. bırakma) tüm listelerin yeni sırayı aynı karede çizeceği bir anda uygulanır: bir sonraki karenin başında, hiçbir zaman bırakıldığı karede değil (`settleDrop`).
- **Önceki karenin `ui.Element`'i kullanılamaz** ("element from pass N used in pass M"). Kareler arası ölçü gerekiyorsa sayıyı sakla (`dropList.heights`), öğeyi değil.
- **Kendi sürükleme katmanımız var.** `Element.Drag` kopyası ve `ui.List` Reorder'ın mavi çizgisi özelleştirilemiyor; bunun yerine `Dragged`/`PointerPosition` ile işaretçi izlenir, kopya `ui.Overlay`'de çizilir. İşaretçi konumu deltaları toplayarak değil `LayoutBounds + PointerPosition` ile okunur.
- **Sadece boyayan animasyon** `Draw` içinde `p.Now()` ve `p.AnimationFrame()` ile yapılır (görünümü yeniden kurmaz): bulut parlaması, spinner.

## 4. Hareketi doğrulamak

Testlerin geçmesi hissin doğru olduğunu kanıtlamaz.

- **Kare kare ölç.** Geçici bir `gallery/zz_*_tmp_test.go` ile `ui.Tester`'da hareketi oynat, her karede `tt.Find(...)` konumunu logla ya da `tt.Image()`'i PNG kaydet; iş bitince geçici dosyayı sil. Sürekliliğe bak: bırakma anında konum sıçramamalı.
- **Açık, koyu ve bevel** üçünde ekran görüntüsü al.
- **Tester gerçek uygulama değildir.** Kareleri farklı sırada/geçişte işleyebilir; bazı hatalar yalnız gerçek pencerede çıkar (bırakma zamanlaması böyle bulundu). Kullanıcı ekran kaydı gönderirse `ffmpeg` ile kare çıkar (`fps=30,crop=…,tile=6x5`) ve olayı karede bul.
- Gerçek pencerede deneyemediğin bir şeyi "düzeldi" diye bildirme; ne doğruladığını, neyi doğrulayamadığını söyle.

## 5. Öneri getir

Kullanıcı her şeyi kendisi düşünmek zorunda kalmasın. Bir bileşene dokunurken ya da iş bitince, ona yakışacak bir mikro etkileşim fırsatı görürsen bir iki cümleyle öner ("şunu da ekleyebiliriz: …"). Öneri somut olsun (ne, nerede, nasıl bir his), bir seferde en çok iki-üç tane, ve yapmadan önce onay al.

Bekleyen fikirler (yapıldıkça sil):

- Panoya kopyala düğmesi: ikon onay işaretine dönüşsün, 1.5 s sonra geri dönsün.
- Tabs: içerik, sekme değişiminin yönünden kayarak gelsin.
- Toast: sağa kaydırarak kapatma (eşik altında lastik gibi geri dönsün).
- Form gönderiminde geçersiz alan hafifçe yatay sallansın, odak ona gitsin.
- Progress %100'e varınca çubuğun üstünden bir parıltı geçsin.
- Sortable ve board: kenara yaklaşınca liste kendiliğinden kaysın.
