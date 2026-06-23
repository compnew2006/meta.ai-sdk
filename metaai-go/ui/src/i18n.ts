import { ref, computed } from 'vue';

export type Locale = 'en' | 'ar';

export const currentLocale = ref<Locale>('en');

// Auto-detect browser language
const browserLang = navigator.language || (navigator as any).userLanguage || 'en';
if (browserLang.startsWith('ar')) {
  currentLocale.value = 'ar';
}

export const translations = {
  en: {
    title: 'Meta AI Dashboard',
    chat: 'Chat',
    stream: 'Stream',
    analyze: 'Analyze Media',
    imageGen: 'Generate Image',
    videoGen: 'Generate Video',
    tokenRequired: 'API Key Required',
    enterToken: 'Enter your API key / Bearer token to continue',
    tokenPlaceholder: 'Bearer token or API Key...',
    save: 'Save & Continue',
    promptPlaceholder: 'Ask Meta AI anything...',
    send: 'Send',
    thinking: 'Thinking mode',
    instant: 'Instant response',
    newConv: 'New Conversation',
    uploadMedia: 'Upload Image to Analyze',
    uploadSuccess: 'Uploaded successfully!',
    uploadError: 'Upload failed',
    mediaId: 'Media ID',
    dragDrop: 'Drag and drop an image here or click to select',
    maxSize: 'Max size: 32MB (PNG, JPEG, GIF)',
    analyzing: 'Analyzing...',
    analyzeQuestionPlaceholder: 'What do you want to ask about this image?',
    generate: 'Generate',
    generating: 'Generating...',
    orientation: 'Orientation',
    square: 'Square',
    landscape: 'Landscape',
    vertical: 'Vertical',
    extendVideo: 'Extend Video (+5s)',
    extending: 'Extending...',
    jobId: 'Job ID',
    status: 'Status',
    queued: 'Queued',
    running: 'Running',
    completed: 'Completed',
    failed: 'Failed',
    logout: 'Logout',
    error: 'Error',
    noMedia: 'No media selected',
    apiError: 'API Error',
    copy: 'Copy',
    copied: 'Copied!',
    messages: 'Messages',
    prompt: 'Prompt',
    results: 'Results',
    submit: 'Submit',
    newChat: 'New chat',
    conversations: 'Conversations',
    rename: 'Rename',
    delete: 'Delete',
    confirmDelete: 'Delete this conversation?',
    emptyChat: 'No messages yet. Say hello!',
    settings: 'Settings',
    systemInstructions: 'System Instructions',
    systemInstructionsPlaceholder: 'e.g. Always answer in Arabic. Be concise and friendly.',
    systemInstructionsGlobalHint: 'Default for new conversations. Each conversation can override this from its menu (⋯).',
    saveSettings: 'Save',
    newAnalysis: 'New analysis',
    uploadForAnalysis: 'Upload an image to start',
    cancel: 'Cancel',
  },
  ar: {
    title: 'لوحة تحكم Meta AI',
    chat: 'محادثة',
    stream: 'بث مباشر',
    analyze: 'تحليل الوسائط',
    imageGen: 'توليد الصور',
    videoGen: 'توليد الفيديو',
    tokenRequired: 'مفتاح واجهة برمجة التطبيقات (API) مطلوب',
    enterToken: 'أدخل مفتاح واجهة برمجة التطبيقات / رمز الحامل للمتابعة',
    tokenPlaceholder: 'رمز الحامل أو مفتاح API...',
    save: 'حفظ ومتابعة',
    promptPlaceholder: 'اسأل Meta AI أي شيء...',
    send: 'إرسال',
    thinking: 'وضع التفكير',
    instant: 'رد فوري',
    newConv: 'محادثة جديدة',
    uploadMedia: 'ارفع صورة لتحليلها',
    uploadSuccess: 'تم الرفع بنجاح!',
    uploadError: 'فشل الرفع',
    mediaId: 'معرف الوسائط',
    dragDrop: 'اسحب وأسقط الصورة هنا أو انقر للاختيار',
    maxSize: 'الحد الأقصى للحجم: 32 ميجابايت (PNG, JPEG, GIF)',
    analyzing: 'جاري التحليل...',
    analyzeQuestionPlaceholder: 'ماذا تريد أن تسأل عن هذه الصورة؟',
    generate: 'توليد',
    generating: 'جاري التوليد...',
    orientation: 'الاتجاه',
    square: 'مربع',
    landscape: 'أفقي',
    vertical: 'عمودي',
    extendVideo: 'تمديد الفيديو (+5 ثوانٍ)',
    extending: 'جاري التمديد...',
    jobId: 'معرف المهمة',
    status: 'الحالة',
    queued: 'في الانتظار',
    running: 'قيد التشغيل',
    completed: 'مكتمل',
    failed: 'فشل',
    logout: 'تسجيل الخروج',
    error: 'خطأ',
    noMedia: 'لم يتم اختيار وسائط',
    apiError: 'خطأ في واجهة برمجة التطبيقات',
    copy: 'نسخ',
    copied: 'تم النسخ!',
    messages: 'الرسائل',
    prompt: 'المطالبة',
    results: 'النتائج',
    submit: 'إرسال',
    newChat: 'محادثة جديدة',
    conversations: 'المحادثات',
    rename: 'إعادة تسمية',
    delete: 'حذف',
    confirmDelete: 'حذف هذه المحادثة؟',
    emptyChat: 'لا توجد رسائل بعد. قل مرحبًا!',
    settings: 'الإعدادات',
    systemInstructions: 'تعليمات النظام',
    systemInstructionsPlaceholder: 'مثال: أجب دائمًا بالعربية. كن موجزًا وودودًا.',
    systemInstructionsGlobalHint: 'افتراضي للمحادثات الجديدة. كل محادثة يمكنها تجاوزه من قائمتها (⋯).',
    saveSettings: 'حفظ',
    newAnalysis: 'تحليل جديد',
    uploadForAnalysis: 'ارفع صورة للبدء',
    cancel: 'إلغاء',
  }
};

export const t = computed(() => translations[currentLocale.value]);

export function toggleLocale() {
  currentLocale.value = currentLocale.value === 'en' ? 'ar' : 'en';
  document.documentElement.dir = currentLocale.value === 'ar' ? 'rtl' : 'ltr';
  document.documentElement.lang = currentLocale.value;
}

// Initial dir setup
if (typeof document !== 'undefined') {
  document.documentElement.dir = currentLocale.value === 'ar' ? 'rtl' : 'ltr';
  document.documentElement.lang = currentLocale.value;
}
