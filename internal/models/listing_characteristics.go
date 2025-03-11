package models

// Константы для типов характеристик
const (
	// Характеристики для цены
	CHAR_PRICE = "price"

	// Характеристики для цвета
	CHAR_COLOR = "color"

	// Характеристики для выпадающего списка
	CHAR_SCREEN_RESOLUTION = "screen_resolution" // Разрешение экрана
	CHAR_DISPLAY_TYPE      = "display_type"      // Тип дисплея

	// Характеристики для селектора
	CHAR_QUALITY = "quality" // Качество
	CHAR_CLASS   = "class"   // Класс

	// Характеристики для чекбокса
	CHAR_IS_NEW        = "is_new"        // Новый товар
	CHAR_HAS_WARRANTY  = "has_warranty"  // Есть гарантия
	CHAR_FAST_DELIVERY = "fast_delivery" // Быстрая доставка

	// Характеристики для размеров
	CHAR_SCREEN_SIZE = "screen_size" // Размер экрана
	CHAR_WEIGHT      = "weight"      // Вес
	CHAR_DIMENSIONS  = "dimensions"  // Габариты
)

// Константы единиц измерений для различных физических величин
const (
	// Длина
	CM = "cm" // сантиметр
	M  = "m"  // метр
	KM = "km" // километр

	// Площадь
	CM2 = "cm2" // квадратный сантиметр
	M2  = "m2"  // квадратный метр
	KM2 = "km2" // квадратный километр

	// Объем
	CM3 = "cm3" // кубический сантиметр
	M3  = "m3"  // кубический метр
	KM3 = "km3" // кубический километр
	ML  = "ml"  // миллилитр
	L   = "l"   // литр

	// Масса
	G  = "g"  // грамм
	KG = "kg" // килограмм
	T  = "t"  // тонна

	// Электричество
	MA = "ma" // миллиампер
	A  = "a"  // ампер
	W  = "w"  // ватт
	KW = "kw" // киловатт
	OM = "om" // ом
)
