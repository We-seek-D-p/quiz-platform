from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy import Sequence
from sqlalchemy.ext.asyncio.session import AsyncSession

from quiz_management.models.question import Question, QuestionCreate, QuestionOption, QuestionUpdate
from quiz_management.repositories.question_repository import QuestionRepository


class QuestionService:
    def __init__(self, db: AsyncSession):
        self.repository = QuestionRepository(db)

    async def get_quiz_questions(self, quiz_id: UUID) -> Sequence[Question]:
        return await self.repository.get_by_quiz_id(quiz_id)

    async def create_question(self, data: QuestionCreate, quiz_id: UUID) -> Question:
        options_data = sorted(data.options, key=lambda x: x.order_index)
        question = Question(**data.model_dump(exclude={"options"}), quiz_id=quiz_id)
        question.options = [
            QuestionOption(**opt.model_dump(exclude={"order_index"}), order_index=idx)
            for idx, opt in enumerate(options_data)
        ]
        return await self.repository.save(question)

    async def update_question(self, question: Question, data: QuestionUpdate) -> Question:
        data_to_update = data.model_dump(exclude_unset=True, exclude={"options"})
        for key, value in data_to_update.items():
            setattr(question, key, value)
        if data.options is not None:
            data.options = sorted(data.options, key=lambda x: x.order_index)

            current_options = {opt.id: opt for opt in question.options if opt.deleted_at is None}
            incoming_ids = {opt.id for opt in data.options if opt.id is not None}

            for option_id, option in current_options.items():
                if option_id not in incoming_ids:
                    option.deleted_at = datetime.now(UTC)

            for order, option in enumerate(data.options):
                if option.id and option.id in current_options:
                    curr_option = current_options[option.id]
                    for key, value in option.model_dump(
                        exclude_unset=True, exclude={"id", "order_index"}
                    ).items():
                        setattr(curr_option, key, value)
                    curr_option.order_index = order
                    curr_option.updated_at = datetime.now(UTC)
                else:
                    new_option = QuestionOption(
                        **option.model_dump(exclude={"id", "order_index"}), order_index=order
                    )
                    question.options.append(new_option)

        if data_to_update or data.options is not None:
            question.updated_at = datetime.now(UTC)

        return await self.repository.save(question)

    async def delete_question(self, question: Question) -> None:
        question.deleted_at = datetime.now(UTC)

        for option in question.options:
            option.deleted_at = datetime.now(UTC)

        await self.repository.save(question)
