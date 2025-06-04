import os
from dotenv import load_dotenv

# Load environment variables from .env file if it exists (for local development)
# This allows developers to set their API key in a local .env file without
# committing it to the repository.
load_dotenv()

DEEPGRAM_API_KEY = os.getenv("DEEPGRAM_API_KEY")

if not DEEPGRAM_API_KEY:
    # This warning is helpful for developers running the service locally.
    # In a production environment, the DEEPGRAM_API_KEY should be set directly
    # as an environment variable.
    print("Warning: DEEPGRAM_API_KEY environment variable not set. SpeechToTextService may not function correctly.")
    # Optionally, you could raise an error here if the API key is absolutely critical
    # for the service to even start, or provide a non-functional default.
    # For placeholder/testing, a default might be okay, but not for production.
    # Example:
    # DEEPGRAM_API_KEY = "YOUR_FALLBACK_KEY_IF_ANY_BUT_USUALLY_RAISE_ERROR"
    # raise ValueError("DEEPGRAM_API_KEY not found in environment variables.")

# Other configurations for the STT service can be added here, for example:
# DEFAULT_LANGUAGE = "en-US"
# DEFAULT_MODEL = "nova-2"
# ENABLE_FORMATTING = True
# ENABLE_PUNCTUATION = True
