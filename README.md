redis cho wiki
[1]. wiki_word:{wiki_word_id}
{
    "id": "uuid of wiki_words",
    "word": "run",
    "pronunciation": "/rʌn/",
    "definitions": [
        {
			"id": "uuid of wiki_word_definition",
            "means": ["di chuyển", "chạy"],
            "part_of_speech": "verb",
            "samples": [
                {
					"id": "uuid of wiki_word_definition_samples",
                    "sample_sentence": "The train runs between Hanoi and Ho Chi Minh City.",
                    "sample_sentence_mean": "Tàu chạy giữa Hà Nội và thành phố Hồ Chí Minh"
                }
            ],
			"synonyms": [
				{
					"id": "uuid of wiki_word_synonyms",
					"wiki_synonym_id": "uuid of wiki_words"
				}
			],
			"antonyms": [
				{
					"id": "uuid of wiki_word_synonyms",
					"wiki_synonym_id": "uuid of wiki_words"
				}
			],
            "is_main_definition": true
        }
    ],  
}

[2]. wiki_phrase:{wiki_phrase_id}
{
    "id": "uuid of wiki_phrases",
    "phrase": "give up",
    "definitions": [
        {
            "id": "uuid of wiki_phrase_definition",
            "mean": ["từ bỏ, "bỏ cuộc"],
            "samples": [
                {
                    "id": "uuid of wiki_phrase_definition_samples",
                    "sample_sentence": "Never give up on your dreams.",
                    "sample_sentence_mean": "Đừng bao giờ từ bỏ ước mơ của bạn."
                }
            ]
            "is_main_definition": true,
        }
    ]
}