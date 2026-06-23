# 🤖 مكتبة Meta AI للغة بايثون (Python SDK)

> **🌐 اللغات:** [🇬🇧 English](README.md) · [🇪🇬 العربية](README.ar.md) (هل تبحث عن النسخة الإنجليزية؟ راجع [README.md](README.md))

[![Python Version](https://img.shields.io/pypi/pyversions/metaai-sdk)](https://pypi.org/project/metaai-sdk/)
[![License](https://img.shields.io/github/license/mir-ashiq/metaai-api)](https://github.com/mir-ashiq/metaai-api/blob/main/LICENSE)
[![PyPI](https://img.shields.io/pypi/v/metaai-sdk)](https://pypi.org/project/metaai-sdk/)
[![GitHub stars](https://img.shields.io/github/stars/mir-ashiq/metaai-api)](https://github.com/mir-ashiq/metaai-api)

أطلق العنان لقوة Meta AI باستخدام بايثون 🚀

مكتبة بايثون (Python SDK) حديثة وغنية بالميزات توفر وصولاً سلسًا إلى إمكانيات Meta AI المتطورة: الدردشة مع Llama 3، توليد الصور، وإنشاء مقاطع الفيديو بالذكاء الاصطناعي - كل ذلك بدون الحاجة لمفاتيح API!

🎯 [البدء السريع](#-البدء-السريع) • 📖 [التوثيق](#-التوثيق) • 💡 [أمثلة](#-أمثلة) • 🎬 [توليد الفيديو](#-توليد-الفيديو)

## ✨ لماذا تختار هذه المكتبة؟

* **🎯 صفر إعدادات (Zero Configuration)**
  لا حاجة لمفاتيح API! فقط قم بالتثبيت وابدأ البرمجة مباشرة.
* **⚡ سرعة فائقة**
  محسنة للأداء وتقدم استجابات فورية في الوقت الفعلي.
* **🔥 كاملة الميزات**
  الدردشة • الصور • الفيديو - كلها في مكتبة واحدة.

## 🌟 القدرات الأساسية

> [!NOTE]
> **تنويه بالحالة الحالية:**
> ميزات الدردشة، الصور، والفيديو تعمل حالياً باستخدام المصادقة المعتمدة على الكوكيز (Cookie-based authentication) بالإضافة إلى رمز Meta AI OAuth المستخدم بواسطة المكتبة وخادم الـ API.
> يظل توليد الصور والفيديو يعمل بالكامل باستخدام مصادقة الكوكيز البسيطة (يتطلب كوكيز 2 فقط).
> راجع ملف [CHANGELOG.md](CHANGELOG.md) لمعرفة سجل الإصدارات وتحديثات التنفيذ.

| الميزة | الوصف | الحالة |
| :--- | :--- | :--- |
| 💬 **الدردشة الذكية** | مدعومة بـ Llama 3 مع وصول للإنترنت | ✅ تعمل |
| 📤 **رفع الصور** | رفع الصور لتوليدها أو تحليلها | ✅ تعمل |
| 🎨 **توليد الصور** | إنشاء صور مذهلة بالذكاء الاصطناعي | ✅ تعمل |
| 🎬 **توليد الفيديو** | توليد مقاطع فيديو من النصوص أو الصور المرفوعة | ✅ تعمل |
| 🔍 **تحليل الصور** | وصف وتحليل واستخراج المعلومات من الصور | ✅ تعمل |
| 🌐 **بيانات حية** | الحصول على معلومات حديثة عبر تكامل Bing | ✅ تعمل |
| 📚 **توثيق المصادر** | تشمل الإجابات مصادر يمكن التحقق منها | ✅ تعمل |
| 🔄 **دعم البث (Streaming)** | بث الاستجابات في الوقت الفعلي | ✅ تعمل |
| 🔐 **مصادقة الكوكيز** | تستخدم كوكيز الجلسة (بدون رموز معقدة) | ✅ تعمل |
| 🌍 **دعم البروكسي** | توجيه الطلبات عبر البروكسي | ✅ تعمل |

## 📦 التثبيت

### المكتبة فقط (خفيفة الوزن)
لاستخدام Meta AI كمكتبة بايثون:
```bash
pip install metaai-sdk
```

### المكتبة + خادم الـ API
لتشغيلها كخدمة REST API:
```bash
pip install metaai-sdk[api]
```

### من المصدر (From Source)
```bash
git clone https://github.com/mir-ashiq/metaai-api.git
cd metaai-api
pip install -e .          # المكتبة فقط
pip install -e ".[api]"   # المكتبة + خادم الـ API
```

**متطلبات النظام:** Python 3.7+ • اتصال بالإنترنت • هذا كل شيء!

---

## 🚀 البدء السريع

> [!IMPORTANT]
> ميزة الدردشة متاحة الآن. استخدم ميزات الدردشة وتوليد الصور وتوليد الفيديو أدناه.

### مثال 1: توليد الصور (يعمل ✅)

```python
from metaai_api import MetaAI

# التهيئة باستخدام المصادقة المعتمدة على الكوكيز
ai = MetaAI()

# توليد الصور
result = ai.generate_image_new(
    prompt="a beautiful sunset over mountains",
    orientation="LANDSCAPE"  # LANDSCAPE أو VERTICAL أو SQUARE
)

if result["success"]:
    print(f"تم توليد {len(result['image_urls'])} صور:")
    for url in result["image_urls"]:
        print(url)
```

**المخرجات (Output):**
```text
Generated 4 images:
https://scontent-arn2-1.xx.fbcdn.net/o1/v/t0/f2/m421/AQN...
https://scontent-arn2-1.xx.fbcdn.net/o1/v/t0/f2/m421/AQM...
https://scontent-arn2-1.xx.fbcdn.net/o1/v/t0/f2/m421/AQO...
https://scontent-arn2-1.xx.fbcdn.net/o1/v/t0/f2/m421/AQM...
```

### مثال 2: توليد الفيديو (يعمل ✅)

```python
from metaai_api import MetaAI

ai = MetaAI()

# توليد الفيديو (يقوم بالتحقق التلقائي من روابط الفيديو افتراضياً)
result = ai.generate_video_new("waves crashing on a beach at sunset")

if result["success"]:
    print(f"تم توليد {len(result['video_urls'])} فيديوهات:")
    for url in result["video_urls"]:
        print(url)

    # معرفات الوسائط لخطوات سير العمل المتقدمة
    print("Media IDs:", result.get("media_ids", []))
```

**المخرجات (Output):**
```text
Generated 4 videos:
https://scontent.xx.fbcdn.net/o1/v/t6/f2/.../video1.mp4?...
https://scontent.xx.fbcdn.net/o1/v/t6/f2/.../video2.mp4?...
https://scontent.xx.fbcdn.net/o1/v/t6/f2/.../video3.mp4?...
https://scontent.xx.fbcdn.net/o1/v/t6/f2/.../video4.mp4?...

Media IDs: ['956278367576451', '956278364243118', '956278370909784', '956278374243117']
```

**العودة السريعة (بدون انتظار/تحقق تلقائي):**
```python
# لاستجابة أسرع (~17 ثانية)، قم بتعطيل التحقق التلقائي
result = ai.generate_video_new(
    "waves crashing",
    auto_poll=False  # يعود فوراً بمعرف المحادثة
)

if result["success"]:
    print(f"شاهد فيديوهاتك على: https://www.meta.ai/prompt/{result['conversation_id']}")
```

### مثال 3: رفع واستخدام الصور (يعمل ✅)

```python
from metaai_api import MetaAI

ai = MetaAI()

# عملية حسابية معقدة
question = "If I invest $10,000 at 7% annual interest compounded monthly for 5 years, how much will I have?"
response = ai.prompt(question)

print(response["message"])
```

**المخرجات (Output):**
```text
With an initial investment of $10,000 at a 7% annual interest rate compounded monthly
over 5 years, you would have approximately $14,176.25.

Here's the breakdown:
- Principal: $10,000
- Interest Rate: 7% per year (0.583% per month)
- Time: 5 years (60 months)
- Compound Frequency: Monthly
- Total Interest Earned: $4,176.25
- Final Amount: $14,176.25

This calculation uses the compound interest formula: A = P(1 + r/n)^(nt)
```

---

## 🔐 خيارات المصادقة (Authentication Options)

تستخدم المكتبة مصادقة بسيطة قائمة على الكوكيز. الحد الأدنى المطلوب:

```python
from metaai_api import MetaAI

# الكوكيز الأساسية المطلوبة
cookies = {
    "datr": "your_datr_value",
    "ecto_1_sess": "your_ecto_1_sess_value"  # الأهم لتوليد الصور/الفيديو
}

ai = MetaAI(cookies=cookies)
```

كوكيز اختيارية (لتحسين التوافق في بعض المناطق جغرافيًا):
```python
# مجموعة كوكيز أكثر اكتمالاً (موصى بها)
cookies = {
    "datr": "your_datr_value",
    "abra_sess": "your_abra_sess_value",  # اختياري - قد لا يتوفر في بعض الدول
    "ecto_1_sess": "your_ecto_1_sess_value"  # الأهم لتوليد الصور/الفيديو
}

ai = MetaAI(cookies=cookies)
```

البديل: تحميل من متغيرات البيئة (Environment Variables):
```python
import os
from metaai_api import MetaAI

# الكوكيز من ملف .env
ai = MetaAI()  # يتم تحميلها تلقائياً من متغيرات البيئة META_AI_*
```

> [!TIP]
> تم إزالة جلب الرموز (lsd/fb_dtsg). واجهات توليد الصور والفيديو تعمل بشكل مثالي مع كوكيز `datr` + `ecto_1_sess` فقط!

---

## 💬 ميزات الدردشة

### الاستجابات المتدفقة (Streaming)
شاهد ردود الذكاء الاصطناعي تظهر في الوقت الفعلي (مثل ChatGPT):
```python
from metaai_api import MetaAI

ai = MetaAI()

print("🤖 AI: ", end="", flush=True)
for chunk in ai.prompt("Explain quantum computing in simple terms", stream=True):
    print(chunk["message"], end="", flush=True)
print("\n")
```

**المخرجات (Output):**
```text
🤖 AI: Quantum computing is like having a super-powered calculator that can solve
problems in completely new ways. Instead of regular computer bits that are either
0 or 1, quantum computers use "qubits" that can be both 0 and 1 at the same time -
imagine flipping a coin that's both heads and tails until you look at it! This
special ability allows quantum computers to process massive amounts of information
simultaneously, making them incredibly fast for specific tasks like drug discovery,
cryptography, and complex simulations.
```

### سياق المحادثة (Conversation Context)
أجرِ محادثات طبيعية تحتفظ بالسياق:
```python
from metaai_api import MetaAI

ai = MetaAI()

# السؤال الأول
response1 = ai.prompt("What are the three primary colors?")
print("Q1:", response1["message"][:100])

# سؤال تالي (يحافظ على سياق المحادثة)
response2 = ai.prompt("How do you mix them to make purple?")
print("Q2:", response2["message"][:150])

# بدء محادثة جديدة تماماً
response3 = ai.prompt("What's the capital of France?", new_conversation=True)
print("Q3:", response3["message"][:50])
```

**المخرجات (Output):**
```text
Q1: The three primary colors are Red, Blue, and Yellow. These colors cannot be created by mixing...

Q2: To make purple, you mix Red and Blue together. The exact shade of purple depends on the ratio - more red creates a reddish-purple (like magenta)...

Q3: The capital of France is Paris, located in the...
```

### استخدام البروكسي (Proxies)
توجيه الطلبات عبر البروكسي:
```python
from metaai_api import MetaAI

# إعداد البروكسي
proxy = {
    'http': 'http://your-proxy-server:8080',
    'https': 'https://your-proxy-server:8080'
}

ai = MetaAI(proxy=proxy)
response = ai.prompt("Hello from behind a proxy!")
print(response["message"])
```

---

## 🌐 خادم REST API (اختياري)

قم بنشر Meta AI كخدمة REST API! نقاط النهاية للدردشة والصور والفيديو تعمل بالكامل.

> [!NOTE]
> تستخدم الدردشة الآن رمز OAuth المستخرج من Meta AI ويتم تعريض نفس تدفق المكتبة عبر خادم الـ API.

### التثبيت
```bash
pip install metaai-sdk[api]
```

### الإعداد
1. احصل على كوكيز Meta AI (انظر قسم [إعداد الكوكيز](#إعداد-جلب-الكوكيز-الخاصة-بك)).
2. قم بإنشاء ملف `.env`:
   ```env
   META_AI_DATR=your_datr_cookie
   META_AI_ECTO_1_SESS=your_ecto_1_sess_cookie

   # اختياري (موصى به عند توفره)
   META_AI_ABRA_SESS=your_abra_sess_cookie
   ```
3. ابدأ تشغيل الخادم:
   ```bash
   uvicorn metaai_api.api_server:app --host 0.0.0.0 --port 8000
   ```
   يبدأ تشغيل الخادم فوراً (بدون تأخير جلب الرموز مسبقاً).

### نقاط نهاية الـ API (API Endpoints)

| نقطة النهاية | الطريقة (Method) | الوصف | الحالة |
| :--- | :--- | :--- | :--- |
| `/healthz` | GET | فحص حالة الخادم | ✅ يعمل |
| `/upload` | POST | رفع الصور للتوليد | ✅ يعمل |
| `/image` | POST | توليد الصور من النصوص | ✅ يعمل |
| `/video` | POST | توليد الفيديو (ينتظر حتى الاكتمال) | ✅ يعمل |
| `/video/extend` | POST | تمديد الفيديو من معرف الوسائط (Media ID) | ✅ يعمل |
| `/video/async` | POST | بدء توليد فيديو غير متزامن (Async) | ✅ يعمل |
| `/video/jobs/{job_id}` | GET | التحقق من حالة وظيفة الفيديو غير المتزامنة | ✅ يعمل |
| `/chat` | POST | إرسال رسائل الدردشة | ✅ يعمل |

### أمثلة الاستخدام (Working Endpoints)

```python
import requests

BASE_URL = "http://localhost:8000"

# فحص حالة الخادم
response = requests.get(f"{BASE_URL}/healthz")
print(response.json())  # {"status": "ok"}

# توليد الصور
images = requests.post(f"{BASE_URL}/image", json={
    "prompt": "Cyberpunk cityscape at night",
    "orientation": "LANDSCAPE"  # LANDSCAPE أو VERTICAL أو SQUARE
}, timeout=200)
result = images.json()
if result["success"]:
    for url in result["image_urls"]:
        print(url)

# توليد الفيديو (متزامن)
video = requests.post(f"{BASE_URL}/video", json={
    "prompt": "waves crashing on beach"
}, timeout=400)
result = video.json()
if result["success"]:
    print("Video URLs:", result.get("video_urls", []))
    print("Media IDs:", result.get("media_ids", []))

# تمديد الفيديو من معرف الوسائط
extended = requests.post(f"{BASE_URL}/video/extend", json={
    "media_id": result["media_ids"][0]
}, timeout=400)
extend_result = extended.json()
if extend_result["success"]:
    print("Extended URLs:", extend_result.get("video_urls", []))
    print("Extended Media IDs:", extend_result.get("media_ids", []))

# توليد فيديو غير متزامن
job = requests.post(f"{BASE_URL}/video/async", json={
    "prompt": "sunset over ocean"
})
job_id = job.json()["job_id"]

# التحقق من الحالة دورياً
import time
while True:
    status = requests.get(f"{BASE_URL}/video/jobs/{job_id}")
    data = status.json()
    if data["status"] == "completed":
        print("Video URLs:", data["result"]["video_urls"])
        break
    time.sleep(5)
```

### الأداء (Performance)
- **توليد الصور:** ~دقيقتين (يرجع 4 صور)
- **توليد الفيديو:** ~40-60 ثانية (يرجع 3-4 فيديوهات)
- **رفع الصور:** أقل من 5 ثوانٍ

### اختبار جميع الميزات (Test All Features)
لتشغيل سيناريوهات الاختبار بالترتيب المناسب:

1. **اختبار تدفق الدردشة:**
   ```bash
   python scripts/test_chat_feature.py --test-api --base-url http://127.0.0.1:8001 --output tests/integration/outputs/chat_feature_test_results.json
   ```
2. **رفع الصور وتوليد الصور وتوليد الرسوم المتحركة:**
   ```bash
   python scripts/test_upload_and_generation.py --base-url http://127.0.0.1:8001
   ```
3. **التحقق الكامل من المكتبة وخادم الـ API:**
   ```bash
   python scripts/test_all_features_complete.py --base-url http://127.0.0.1:8001 --output tests/integration/outputs/feature_test_report_sdk_api_final.json
   ```

*أضف `--video-auto-poll` إذا كنت تريد تشغيل الانتظار التلقائي لروابط الوسائط النهائية أثناء التحقق من الفيديو.*

### سير عمل تمديد الرسوم المتحركة + رفع الصور (SDK)

```python
from metaai_api import MetaAI

ai = MetaAI()

# 1) رفع صورة
upload = ai.upload_image("path/to/image.jpg")
if not upload.get("success"):
    raise RuntimeError("Upload failed")

media_id = upload["media_id"]
metadata = {
    "file_size": upload.get("file_size", 0),
    "mime_type": upload.get("mime_type", "image/jpeg"),
}

# 2) تحريك الصورة المرفوعة إلى فيديو
video = ai.generate_video_new(
    prompt="animate this image with smooth cinematic motion",
    media_ids=[media_id],
    attachment_metadata=metadata,
)

# 3) تمديد مقطع فيديو متحرك تم توليده
if video.get("success") and video.get("media_ids"):
    extended = ai.extend_video(video["media_ids"][0])
    print("Extended video URLs:", extended.get("video_urls", []))
```

---

## 🎬 توليد الفيديو

إنشاء مقاطع فيديو مولدة بالذكاء الاصطناعي من الأوصاف النصية!

### إعداد: جلب الكوكيز الخاصة بك
1. قم بزيارة `meta.ai` في متصفحك وسجل الدخول.
2. افتح DevTools (F12) ← علامة تبويب **Application** ← الكوكيز (**Cookies**) ← `https://meta.ai`.
3. انسخ قيم الكوكيز المطلوبة التالية:
   * `datr`
   * `ecto_1_sess` (الأهم للتوليد)
4. اختياري (إن وجد):
   * `abra_sess`

> [!NOTE]
> كوكيز `datr` و `ecto_1_sess` هي المطلوبة فقط. لا حاجة لرموز أخرى (`lsd`/`fb_dtsg`)!

### 🔄 التحديث التلقائي للكوكيز
تنتهي صلاحية الكوكيز (خاصة `ecto_1_sess`) بشكل دوري. تتضمن المكتبة الآن سكريبتات تحديث تلقائي للكوكيز!

#### الخيار 1: التصدير اليدوي (موصى به)
1. في متصفحك: **Copy as cURL** ← واحفظه كملف `curl.json`
2. قم بتشغيل المستخرج:
   ```bash
   python refresh_cookies.py
   ```

#### متى يجب التحديث؟
تكتشف المكتبة تلقائياً الكوكيز منتهية الصلاحية وستظهر:
```text
❌ Cookie Expired: ecto_1_sess needs to be refreshed
Run: python auto_refresh_cookies.py
```

**الكوكيز الأساسية:**
* `ecto_1_sess` ⭐ - رمز الجلسة (ينتهي بشكل متكرر، ويجب تحديثه)
* `rd_challenge` - كوكيز التحدي (يتم تحديثه تلقائياً بواسطة المكتبة)
* `ps_l`, `ps_n` - علامات البوابة (اختيارية، قد تحسن الموثوقية)

### مثال 1: توليد أول فيديو لك

```python
from metaai_api import MetaAI

# كوكيز المتصفح الخاص بك (الحد الأدنى المطلوب)
cookies = {
    "datr": "your_datr_value_here",
    "ecto_1_sess": "your_ecto_1_sess_value_here"
}

# التهيئة
ai = MetaAI(cookies=cookies)

# توليد فيديو
result = ai.generate_video_new("A majestic lion walking through the African savanna at sunset")

if result["status"] == "READY":
    print("✅ Video generated successfully!")
    print(f"🎬 Generated {len(result['video_urls'])} videos")
    for i, url in enumerate(result['video_urls'], 1):
        print(f"   Video {i}: {url[:80]}...")
    print(f"📝 Prompt: {result['prompt']}")
elif result["status"] == "PROCESSING":
    print("⏳ Video request accepted and still processing")
    print("Media IDs:", result.get("media_ids", []))
else:
    print("❌ Video generation failed")
    print(result.get("error"))
    print(result.get("graphql_errors", []))
```

**المخرجات (Output):**
```text
✅ Sending video generation request...
✅ Video generation request sent successfully!
⏳ Waiting before polling...
🔄 Polling for video URLs (Attempt 1/20)...
✅ Video URLs found!

✅ Video generated successfully!
🎬 Generated 3 videos
   Video 1: https://scontent.xx.fbcdn.net/v/t66.36240-6/video1.mp4?...
   Video 2: https://scontent.xx.fbcdn.net/v/t66.36240-6/video2.mp4?...
   Video 3: https://scontent.xx.fbcdn.net/v/t66.36240-6/video3.mp4?...
📝 Prompt: A majestic lion walking through the African savanna at sunset
```

### كيفية جلب الكوكيز الخاصة بك
1. افتح `https://meta.ai` في متصفحك وسجل الدخول.
2. اضغط **F12** ← علامة تبويب **Application**.
3. توجه إلى **Cookies** ← `https://meta.ai`.
4. انسخ هذه القيم المطلوبة:
   * `datr`
   * `ecto_1_sess`
   * *اختياري:* `abra_sess`
5. أضفها إلى كود بايثون أو ملف `.env`.

### مثال 2: توليد فيديوهات متعددة

```python
from metaai_api import MetaAI
import time

ai = MetaAI(cookies=cookies)

prompts = [
    "A futuristic city with flying cars at night",
    "Ocean waves crashing on a tropical beach",
    "Northern lights dancing over a snowy mountain"
]

videos = []
for i, prompt in enumerate(prompts, 1):
    print(f"\n🎬 Generating video {i}/{len(prompts)}: {prompt}")
    result = ai.generate_video(prompt, verbose=False)

    if result["success"]:
        videos.append(result["video_urls"][0])
        print(f"✅ Success! URL: {result['video_urls'][0][:50]}...")
    else:
        print("⏳ Still processing...")

    time.sleep(5)  # Be nice to the API

print(f"\n🎉 Generated {len(videos)} videos successfully!")
```

**المخرجات (Output):**
```text
🎬 Generating video 1/3: A futuristic city with flying cars at night
✅ Success! URL: https://scontent.xx.fbcdn.net/v/t66.36240-6/1234...

🎬 Generating video 2/3: Ocean waves crashing on a tropical beach
✅ Success! URL: https://scontent.xx.fbcdn.net/v/t66.36240-6/5678...

🎬 Generating video 3/3: Northern lights dancing over a snowy mountain
✅ Success! URL: https://scontent.xx.fbcdn.net/v/t66.36240-6/9012...

🎉 Generated 3 videos successfully!
```

### مثال 3: توليد فيديو متقدم مع تحديد الاتجاه (Orientation)

```python
from metaai_api import MetaAI

ai = MetaAI(cookies=cookies)

# توليد فيديو مع تحديد الاتجاه (الافتراضي هو VERTICAL)
result = ai.generate_video(
    prompt="A time-lapse of a flower blooming",
    orientation="VERTICAL",   # الخيارات: "LANDSCAPE", "VERTICAL", "SQUARE"
    wait_before_poll=15,      # الانتظار 15 ثانية قبل الفحص الأول
    max_attempts=50,          # محاولة حتى 50 مرة
    wait_seconds=3,           # الانتظار 3 ثوانٍ بين المحاولات
    verbose=True              # إظهار تقدم العملية بالتفصيل
)

# توليد فيديو أفقي للشاشات العريضة
result_landscape = ai.generate_video(
    prompt="Panoramic view of sunset over mountains",
    orientation="LANDSCAPE"   # تنسيق عريض (16:9)
)

if result["success"]:
    print(f"\n🎬 فيديوهاتك جاهزة!")
    print(f"🔗 تم توليد {len(result['video_urls'])} فيديوهات:")
    for i, url in enumerate(result['video_urls'], 1):
        print(f"   Video {i}: {url}")
    print(f"⏱️ تم التوليد في: {result['timestamp']}")
```

**اتجاهات الفيديو المدعومة:**
- `"LANDSCAPE"` - أفقي وعريض (16:9) - مثالي للشاشات العريضة والمحتوى السينمائي.
- `"VERTICAL"` - طولي ورأسي (9:16) - مثالي للهواتف المحمولة والقصص والـ Reels (الافتراضي).
- `"SQUARE"` - أبعاد متساوية (1:1) - مثالي لمنشورات التواصل الاجتماعي.

> 📖 **دليل الفيديو الكامل:** راجع [GENERATION_API.md](GENERATION_API.md) للتوثيق الكامل.

---

## 📤 رفع وتحليل الصور

ارفع الصور إلى Meta AI لتحليلها، وتوليد صور مماثلة، وإنشاء فيديو:

### رفع وتحليل الصور

```python
from metaai_api import MetaAI

# التهيئة باستخدام الكوكيز (datr + ecto_1_sess مطلوبة)
ai = MetaAI(cookies={
    "datr": "your_datr_cookie",
    "ecto_1_sess": "your_ecto_1_sess_cookie",
    # "abra_sess": "your_abra_sess_cookie"  # اختياري
})

# الخطوة 1: رفع صورة
result = ai.upload_image("path/to/image.jpg")

if result["success"]:
    media_id = result["media_id"]
    metadata = {
        'file_size': result['file_size'],
        'mime_type': result['mime_type']
    }

    # الخطوة 2: تحليل الصورة
    response = ai.prompt(
        message="What do you see in this image? Describe it in detail.",
        media_ids=[media_id],
        attachment_metadata=metadata
    )
    print(f"🔍 Analysis: {response['message']}")

    # الخطوة 3: توليد صور مماثلة
    response = ai.prompt(
        message="Create a similar image in watercolor painting style",
        media_ids=[media_id],
        attachment_metadata=metadata,
        is_image_generation=True
    )
    print(f"🎨 Generated {len(response['media'])} similar images")

    # الخطوة 4: توليد فيديو من الصورة
    video = ai.generate_video(
        prompt="generate a video with zoom in effect on this image",
        media_ids=[media_id],
        attachment_metadata=metadata
    )
    if video["success"]:
        print(f"🎬 Video: {video['video_urls'][0]}")
```

**المخرجات (Output):**
```text
🔍 Analysis: The image captures a serene lake scene set against a majestic mountain backdrop. In the foreground, there's a small, golden-yellow wooden boat with a bright yellow canopy floating on calm, glass‑like water...

🎨 Generated 4 similar images

🎬 Video: https://scontent.fsxr1-2.fna.fbcdn.net/o1/v/t6/f2/m421/video.mp4
```
> 📖 **دليل رفع الصور الكامل:** راجع `examples/image_upload_example.py` لمثال عملي على رفع وتحليل الصور.

---

## 🎨 توليد الصور

توليد صور مدعومة بالذكاء الاصطناعي باتجاهات مخصصة (تتطلب مصادقة فيسبوك):

```python
from metaai_api import MetaAI

# التهيئة باستخدام بيانات اعتماد فيسبوك
ai = MetaAI(fb_email="your_email@example.com", fb_password="your_password")

# توليد صور بالاتجاه الافتراضي (VERTICAL)
response = ai.prompt("Generate an image of a cyberpunk cityscape at night with neon lights")

# أو تحديد الاتجاه بصورة صريحة
response_landscape = ai.prompt(
    "Generate an image of a panoramic mountain landscape",
    orientation="LANDSCAPE"  # الخيارات: "LANDSCAPE", "VERTICAL", "SQUARE"
)

response_vertical = ai.prompt(
    "Generate an image of a tall waterfall",
    orientation="VERTICAL"  # تنسيق رأسي/طولي (الافتراضي)
)

response_square = ai.prompt(
    "Generate an image of a centered mandala pattern",
    orientation="SQUARE"  # تنسيق مربع (1:1)
)

# عرض النتائج (تقوم Meta AI بتوليد 4 صور افتراضياً)
print(f"🎨 Generated {len(response['media'])} images:")
for i, image in enumerate(response['media'], 1):
    print(f"  Image {i}: {image['url']}")
    print(f"  Prompt: {image['prompt']}")
```

**الاتجاهات المدعومة:**
- `"LANDSCAPE"` - تنسيق أفقي عريض (16:9) - مثالي للمناظر الطبيعية والبانوراما.
- `"VERTICAL"` - تنسيق طولي رأسي (9:16) - مثالي للبورتريه ومحتوى الهاتف المحمول (الافتراضي).
- `"SQUARE"` - أبعاد متساوية (1:1) - لوسائل التواصل الاجتماعي وصور الملفات الشخصية.

**المخرجات (Output):**
```text
🎨 Generated 4 images:
  Image 1: https://scontent.xx.fbcdn.net/o1/v/t0/f1/m247/img1.jpeg
  Prompt: a cyberpunk cityscape at night with neon lights

  Image 2: https://scontent.xx.fbcdn.net/o1/v/t0/f1/m247/img2.jpeg
  Prompt: a cyberpunk cityscape at night with neon lights

  Image 3: https://scontent.xx.fbcdn.net/o1/v/t0/f1/m247/img3.jpeg
  Prompt: a cyberpunk cityscape at night with neon lights

  Image 4: https://scontent.xx.fbcdn.net/o1/v/t0/f1/m247/img4.jpeg
  Prompt: a cyberpunk cityscape at night with neon lights
```

---

## 💡 أمثلة

استكشف الأمثلة الجاهزة للاستخدام في مجلد `examples/`:

| الملف | الوصف | الميزات |
| :--- | :--- | :--- |
| 📄 `image_workflow_complete.py` | سير عمل كامل للصور | رفع، تحليل، توليد صور وفيديو |
| 📄 `simple_example.py` | دليل البدء السريع | دردشة أساسية + توليد فيديو |
| 📄 `video_generation.py` | توليد الفيديو | أمثلة متعددة ومعالجة الأخطاء |
| 📄 `test_example.py` | حزمة الاختبارات | التحقق والاختبار |

### تشغيل مثال
```bash
# استنساخ المستودع
git clone https://github.com/mir-ashiq/metaai-api.git
cd metaai-api

# تشغيل المثال البسيط
python examples/simple_example.py

# تشغيل أمثلة توليد الفيديو
python examples/video_generation.py
```

---

## 📖 التوثيق

### 📚 الأدلة الكاملة

| المستند | الوصف |
| :--- | :--- |
| 📘 **[البدء السريع](QUICK_START.md)** | إعداد المكتبة والـ API وأول الطلبات |
| 📘 **[واجهة توليد الصور والفيديو](GENERATION_API.md)** | تفاصيل توليد الصور والفيديو |
| 📙 **[التغييرات والكوكيز](CHANGES_AND_COOKIES.md)** | إعداد الكوكيز والتنبيهات المعروفة |
| 📕 **[دليل المساهمة](CONTRIBUTING.md)** | كيفية المساهمة في المشروع |
| 📔 **[سجل التغييرات](CHANGELOG.md)** | تاريخ الإصدارات والتحديثات |
| 📓 **[سياسة الأمان](SECURITY.md)** | أفضل الممارسات الأمنية |

### 🔧 مرجع الـ API

#### كلاس `MetaAI`
```python
class MetaAI:
    def __init__(
        self,
        fb_email: Optional[str] = None,
        fb_password: Optional[str] = None,
        cookies: Optional[dict] = None,
        proxy: Optional[dict] = None
    )
```

**الوظائف (Methods):**
- `prompt(message, stream=False, new_conversation=False)`
  إرسال رسالة دردشة.
  *يرجع:* قاموساً (`dict`) بالنص والمصادر والوسائط.
- `generate_video(prompt, wait_before_poll=10, max_attempts=30, wait_seconds=5, verbose=True)`
  توليد فيديو من النص.
  *يرجع:* قاموساً (`dict`) بالنجاح وروابط الفيديو ومعرفات الوسائط ومُعرّف المحادثة والنص والوقت.
- `extend_video(media_id, source_media_url=None, conversation_id=None, wait_before_poll=10, max_attempts=30, wait_seconds=5, verbose=True)`
  تمديد فيديو مولد سابقاً من معرف الوسائط.
  *يرجع:* قاموساً (`dict`) بالنجاح وروابط الفيديو الممتدة ومعرفات الوسائط.

#### كلاس `VideoGenerator`
```python
from metaai_api import VideoGenerator

# توليد فيديو مباشر
generator = VideoGenerator(cookies_str="your_cookies_as_string")
result = generator.generate_video("your prompt here")

# توليد في سطر واحد
result = VideoGenerator.quick_generate(
    cookies_str="your_cookies",
    prompt="your prompt"
)
```

---

## 🎯 حالات الاستخدام (Use Cases)

### 1. مساعد بحثي (Research Assistant)
```python
ai = MetaAI()
research = ai.prompt("Summarize recent breakthroughs in fusion energy")
print(research["message"])
# الحصول على المصادر
for source in research["sources"]:
    print(f"📌 {source['title']}: {source['link']}")
```

### 2. صناعة المحتوى (Content Creation)
```python
ai = MetaAI(cookies=cookies)

# توليد محتوى الفيديو
promo_video = ai.generate_video("Product showcase with smooth camera movements")

# توليد صور مصغرة
thumbnails = ai.prompt("Generate a YouTube thumbnail for a tech review video")
```

### 3. أداة تعليمية (Educational Tool)
```python
ai = MetaAI()

# شرح مواضيع معقدة
explanation = ai.prompt("Explain blockchain technology to a 10-year-old")

# المساعدة في حل الواجبات
solution = ai.prompt("Solve: 2x + 5 = 13, show steps")
```

### 4. معلومات في الوقت الفعلي (Real-time Information)
```python
ai = MetaAI()

# أحداث جارية
news = ai.prompt("What are the top technology news today?")

# نتائج رياضية
scores = ai.prompt("Latest Premier League scores")

# بيانات السوق
stocks = ai.prompt("Current S&P 500 index value")
```

---

## 🛠️ الإعدادات المتقدمة

### متغيرات البيئة (Environment Variables)
قم بتخزين الكوكيز والبيانات بشكل آمن في ملف `.env`:
```env
META_AI_DATR=your_datr_value
META_AI_ECTO_1_SESS=your_ecto_1_sess_value

# اختياري
META_AI_ABRA_SESS=your_abra_sess_cookie

# اختياري لتجاوز معرفات الاستعلام المستمرة (persisted queries)
META_AI_DOC_ID_TEXT_TO_IMAGE=override_doc_id
META_AI_DOC_ID_TEXT_TO_VIDEO=override_doc_id
# ...أو استخدم META_AI_DOC_ID لكلاهما
```

التحميل في بايثون:
```python
from metaai_api import MetaAI

# يتم التحميل تلقائياً من متغيرات البيئة
ai = MetaAI()

# أو تحميل يدوي باستخدام dotenv
import os
from dotenv import load_dotenv

load_dotenv()
cookies = {
    "datr": os.getenv("META_AI_DATR"),
    "ecto_1_sess": os.getenv("META_AI_ECTO_1_SESS")
}

if os.getenv("META_AI_ABRA_SESS"):
    cookies["abra_sess"] = os.getenv("META_AI_ABRA_SESS")

ai = MetaAI(cookies=cookies)
```

### معالجة الأخطاء (Error Handling)

```python
from metaai_api import MetaAI

ai = MetaAI(cookies=cookies)

try:
    result = ai.generate_video_new("Your prompt")

    if result["status"] == "READY":
        print(f"✅ Video: {result['video_urls'][0]}")
    elif result["status"] == "PROCESSING":
        print("⏳ Video still processing")
        print("Media IDs:", result.get("media_ids", []))
    else:
        print("❌ Video generation failed")
        print(result.get("error"))
        print(result.get("graphql_errors", []))

except ValueError as e:
    print(f"❌ Configuration error: {e}")
except ConnectionError as e:
    print(f"❌ Network error: {e}")
except Exception as e:
    print(f"❌ Unexpected error: {e}")
```

---

## 🌟 هيكل المشروع

```text
metaai-api/
│
├── 📁 src/metaai_api/        # الحزمة الأساسية
│   ├── __init__.py            # تهيئة الحزمة
│   ├── main.py                # كلاس MetaAI
│   ├── video_generation.py    # توليد الفيديو
│   ├── client.py              # أدوات العميل المساعدة
│   ├── utils.py               # دوال مساعدة عامة
│   └── exceptions.py          # الاستثناءات المخصصة
│
├── 📁 examples/               # أمثلة الاستخدام
│   ├── simple_example.py      # بدء سريع
│   ├── video_generation.py    # أمثلة الفيديو
│   └── test_example.py        # اختبارات
│
├── 📁 .github/                # إعدادات غيت هاب
│   ├── workflows/             # مسارات العمل والتحقق
│   └── README.md
│
├── 📄 README.md               # الملف بالإنجليزية
├── 📄 QUICK_START.md
├── 📄 GENERATION_API.md
├── 📄 CHANGES_AND_COOKIES.md
├── 📄 CONTRIBUTING.md
├── 📄 CHANGELOG.md
├── 📄 SECURITY.md
├── 📄 LICENSE                 # رخصة MIT
├── 📄 pyproject.toml          # بيانات تعريف المشروع
└── 📄 requirements.txt        # الاعتمادات البرمجية
```

---

## 🤝 المساهمة

نرحب بمساهماتكم! إليك كيف يمكنك المساعدة:

- **🐛 الإبلاغ عن الأخطاء** - افتح issue جديدة.
- **💡 اقتراح ميزات جديدة** - شاركنا أفكارك.
- **📝 تحسين التوثيق** - ساعدنا في تحسين التوثيق.
- **🔧 إرسال طلبات سحب (PRs)** - لإصلاح المشاكل أو إضافة ميزات جديدة.

راجع ملف [CONTRIBUTING.md](CONTRIBUTING.md) للحصول على إرشادات تفصيلية.

---

## 📜 الترخيص

هذا المشروع مرخص تحت رخصة MIT - انظر ملف [LICENSE](LICENSE) لمزيد من التفاصيل.

### ⚖️ إخلاء المسؤولية
هذا المشروع هو تنفيذ مستقل وليس تابعاً رسمياً لشركة Meta Platforms, Inc. أو أي من الشركات التابعة لها.
- ✅ لأغراض تعليمية وتطويرية فقط.
- ✅ استخدمه بمسؤولية وبشكل أخلاقي.
- ✅ الالتزام بشروط خدمة Meta.
- ✅ احترام حدود الاستخدام والسياسات.

**رخصة Llama 3:** قم بزيارة [llama.com/llama3/license](https://llama.com/llama3/license) لمعرفة شروط استخدام Llama 3.

### 🙏 شكر وتقدير
- **Meta AI** - لتقديم قدرات الذكاء الاصطناعي.
- **Llama 3** - نموذج اللغة القوي.
- **مجتمع البرمجيات الحرة** - للإلهام والدعم.

---

## 📞 الدعم والمجتمع
- **💬 أسئلة؟** [GitHub Discussions](https://github.com/mir-ashiq/metaai-api/discussions)
- **🐛 تقارير الأخطاء:** [GitHub Issues](https://github.com/mir-ashiq/metaai-api/issues)
- **📧 البريد الإلكتروني:** imseldrith@gmail.com
- **⭐ ضع نجمة للمستودع على GitHub**

### 🚀 روابط سريعة

| المورد | الرابط |
| :--- | :--- |
| 📦 **حزمة PyPI** | [pypi.org/project/metaai-sdk](https://pypi.org/project/metaai-sdk/) |
| 🐙 **مستودع GitHub** | [github.com/mir-ashiq/metaai-api](https://github.com/mir-ashiq/metaai-api) |
| 📖 **التوثيق الكامل** | [البدء السريع](QUICK_START.md) • [واجهة التوليد](GENERATION_API.md) |
| 💬 **الحصول على المساعدة** | [المشاكل](https://github.com/mir-ashiq/metaai-api/issues) • [المناقشات](https://github.com/mir-ashiq/metaai-api/discussions) |

---
**Meta AI Python SDK v2.0.0** | صُنع بحب ❤️ بواسطة mir-ashiq | رخصة MIT

⭐ ضع نجمة (Star) لهذا المستودع إذا وجدته مفيداً!
